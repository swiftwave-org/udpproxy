package main

import (
	"fmt"
	"sync"
)

type Manager struct {
	Mutex         sync.RWMutex
	ProxyMap      map[string]*UDPProxy // map of <frontendPort:backendService:backendPort> to UDPProxy
	PortsOccupied map[int]bool
}

type ProxyRecord struct {
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Service    string `json:"service"`
}

func (p *ProxyRecord) String() string {
	return proxyNameFormat(p.Port, p.Service, p.TargetPort)
}

func (p *UDPProxy) toRecord() ProxyRecord {
	return ProxyRecord{
		Port:       p.port,
		TargetPort: p.targetPort,
		Service:    p.service,
	}
}

func UDPProxyToRecordList(proxyMap map[string]*UDPProxy) []ProxyRecord {
	var records []ProxyRecord
	for _, proxy := range proxyMap {
		records = append(records, proxy.toRecord())
	}
	return records
}

type AddProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type RemoveProxyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type IsExistProxyResponse struct {
	Exist bool `json:"exist"`
}

func proxyNameFormat(port int, service string, targetPort int) string {
	return fmt.Sprintf("%d:%s:%d", port, service, targetPort)
}
