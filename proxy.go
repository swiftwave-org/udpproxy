package main

// This has been taken from moby/libnetwork - Apache 2.0 License and modified to fit the needs of this project (swiftwave)
// Ref - https://github.com/moby/libnetwork/blob/master/cmd/proxy/udp_proxy.go

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	DNSError = errors.New("DNS resolution failed")
	DNSNoIP  = errors.New("No IP address found")
)

const (
	// UDPConnTrackTimeout is the timeout used for UDP connection tracking
	UDPConnTrackTimeout = 90 * time.Second
	// UDPBufSize is the buffer size for the UDP proxy
	UDPBufSize = 65507
)

// A net.Addr where the IP is split into two fields so you can use it as a key
// in a map:
type connTrackKey struct {
	IPHigh uint64
	IPLow  uint64
	Port   int
}

func newConnTrackKey(addr *net.UDPAddr) *connTrackKey {
	if len(addr.IP) == net.IPv4len {
		return &connTrackKey{
			IPHigh: 0,
			IPLow:  uint64(binary.BigEndian.Uint32(addr.IP)),
			Port:   addr.Port,
		}
	}
	return &connTrackKey{
		IPHigh: binary.BigEndian.Uint64(addr.IP[:8]),
		IPLow:  binary.BigEndian.Uint64(addr.IP[8:]),
		Port:   addr.Port,
	}
}

type connTrackMap map[connTrackKey]*net.UDPConn

// UDPProxy is proxy for which handles UDP datagrams. It implements the Proxy
// interface to handle UDP traffic forwarding between the frontend and backend
// addresses.
type UDPProxy struct {
	listener       *net.UDPConn
	port           int
	targetPort     int
	serviceName    string
	connTrackTable connTrackMap
	connTrackLock  sync.Mutex
}

// NewUDPProxy creates a new UDPProxy.
func NewUDPProxy(port int, targetPort int, serviceName string) (*UDPProxy, error) {
	// detect version of hostIP to bind only to correct version
	listener, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		return nil, err
	}
	return &UDPProxy{
		listener:       listener,
		port:           port,
		targetPort:     targetPort,
		serviceName:    serviceName,
		connTrackTable: make(connTrackMap),
	}, nil
}

func (proxy *UDPProxy) replyLoop(proxyConn *net.UDPConn, clientAddr *net.UDPAddr, clientKey *connTrackKey) {
	defer func() {
		proxy.connTrackLock.Lock()
		delete(proxy.connTrackTable, *clientKey)
		proxy.connTrackLock.Unlock()
		proxyConn.Close()
	}()

	readBuf := make([]byte, UDPBufSize)
	for {
		proxyConn.SetReadDeadline(time.Now().Add(UDPConnTrackTimeout))
	again:
		read, err := proxyConn.Read(readBuf)
		if err != nil {
			if err, ok := err.(*net.OpError); ok && err.Err == syscall.ECONNREFUSED {
				// This will happen if the last write failed
				// (e.g: nothing is actually listening on the
				// proxied port on the container), ignore it
				// and continue until UDPConnTrackTimeout
				// expires:
				goto again
			}
			return
		}
		for i := 0; i != read; {
			written, err := proxy.listener.WriteToUDP(readBuf[i:read], clientAddr)
			if err != nil {
				return
			}
			i += written
		}
	}
}

// Run starts forwarding the traffic using UDP.
func (proxy *UDPProxy) Run() {
	readBuf := make([]byte, UDPBufSize)
	for {
		read, from, err := proxy.listener.ReadFromUDP(readBuf)
		if err != nil {
			// NOTE: Apparently ReadFrom doesn't return
			// ECONNREFUSED like Read do (see comment in
			// UDPProxy.replyLoop)
			if !isClosedError(err) {
				log.Printf("Stopping proxy on port/%d for service %s udp/%d (%s)", proxy.port, proxy.serviceName, proxy.targetPort, err)
			}
			break
		}

		fromKey := newConnTrackKey(from)
		proxy.connTrackLock.Lock()
		proxyConn, hit := proxy.connTrackTable[*fromKey]
		if !hit {
			ip, err := resolveDNS(proxy.serviceName)
			if err != nil {
				// TODO:  report to swiftwave that no IP address found
				log.Printf("Can't resolve the DNS for %s: %s\n", proxy.serviceName, err)
				proxy.connTrackLock.Unlock()
				continue
			}

			backendAddr := &net.UDPAddr{
				IP:   net.ParseIP(ip),
				Port: proxy.targetPort,
				Zone: "",
			}
			proxyConn, err = net.DialUDP("udp", nil, backendAddr)
			if err != nil {
				log.Printf("Can't proxy a datagram to %s udp/%d: %s\n", proxy.serviceName, proxy.targetPort, err)
				proxy.connTrackLock.Unlock()
				continue
			}
			proxy.connTrackTable[*fromKey] = proxyConn
			go proxy.replyLoop(proxyConn, from, fromKey)
		}
		proxy.connTrackLock.Unlock()
		for i := 0; i != read; {
			written, err := proxyConn.Write(readBuf[i:read])
			if err != nil {
				log.Printf("Can't proxy a datagram to %s udp/%d: %s\n", proxy.serviceName, proxy.targetPort, err)
				break
			}
			i += written
		}
	}
}

// Close stops forwarding the traffic.
func (proxy *UDPProxy) Close() {
	proxy.listener.Close()
	proxy.connTrackLock.Lock()
	defer proxy.connTrackLock.Unlock()
	for _, conn := range proxy.connTrackTable {
		conn.Close()
	}
}

func isClosedError(err error) bool {
	/* This comparison is ugly, but unfortunately, net.go doesn't export errClosing.
	 * See:
	 * http://golang.org/src/pkg/net/net.go
	 * https://code.google.com/p/go/issues/detail?id=4337
	 * https://groups.google.com/forum/#!msg/golang-nuts/0_aaCvBmOcM/SptmDyX1XJMJ
	 */
	return strings.HasSuffix(err.Error(), "use of closed network connection")
}

func resolveDNS(serviceName string) (string, error) {
	/*
	 * Resolve the DNS for the service name
	 * Use the system default DNS resolver
	 * Return only single IP address
	 */
	ips, err := net.LookupIP(serviceName)
	if err != nil {
		return "", DNSError
	}
	if len(ips) == 0 {
		return "", DNSNoIP
	}
	return ips[0].String(), nil
}
