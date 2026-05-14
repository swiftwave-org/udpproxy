package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// SwiftwaveCallbackURL is the optional endpoint to which udpproxy POSTs
// failure notifications. When empty, notifications are skipped.
var SwiftwaveCallbackURL = os.Getenv("SWIFTWAVE_CALLBACK_URL")

// dnsFailureDedupWindow throttles repeated notifications for the same service
// so a continuously failing resolver doesn't flood swiftwave.
const dnsFailureDedupWindow = 5 * time.Minute

var (
	notifyClient = &http.Client{Timeout: 5 * time.Second}

	lastDNSFailureNotified = make(map[string]time.Time)
	lastDNSFailureLock     sync.Mutex
)

// DNSFailureReport is the payload posted to SwiftwaveCallbackURL when a proxy
// cannot resolve its backend service.
type DNSFailureReport struct {
	Event      string `json:"event"`
	Service    string `json:"service"`
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Error      string `json:"error"`
	Timestamp  int64  `json:"timestamp"`
}

// reportDNSFailure notifies swiftwave that DNS resolution failed for service.
// No-ops when SwiftwaveCallbackURL is unset and is deduplicated per service
// within dnsFailureDedupWindow. Runs the HTTP call in a goroutine so it never
// blocks the proxy datagram loop.
func reportDNSFailure(service string, port, targetPort int, cause error) {
	if SwiftwaveCallbackURL == "" {
		return
	}

	lastDNSFailureLock.Lock()
	if t, ok := lastDNSFailureNotified[service]; ok && time.Since(t) < dnsFailureDedupWindow {
		lastDNSFailureLock.Unlock()
		return
	}
	lastDNSFailureNotified[service] = time.Now()
	lastDNSFailureLock.Unlock()

	report := DNSFailureReport{
		Event:      "dns_resolution_failed",
		Service:    service,
		Port:       port,
		TargetPort: targetPort,
		Error:      cause.Error(),
		Timestamp:  time.Now().Unix(),
	}

	go func() {
		body, err := json.Marshal(report)
		if err != nil {
			log.Printf("notifier: failed to marshal DNS failure report: %s\n", err)
			return
		}
		req, err := http.NewRequest(http.MethodPost, SwiftwaveCallbackURL, bytes.NewReader(body))
		if err != nil {
			log.Printf("notifier: failed to build DNS failure request: %s\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := notifyClient.Do(req)
		if err != nil {
			log.Printf("notifier: failed to POST DNS failure to swiftwave (%s): %s\n", SwiftwaveCallbackURL, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Printf("notifier: swiftwave returned %d for DNS failure report on service %s\n", resp.StatusCode, service)
			return
		}
		fmt.Printf("notifier: reported DNS failure for service %s to swiftwave\n", service)
	}()
}
