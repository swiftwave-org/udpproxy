package main

import (
	"fmt"
	"sync"
)

type Manager struct {
	Mutex    sync.RWMutex
	ProxyMap map[string]*UDPProxy // map of <frontendPort:backendService:backendPort> to UDPProxy
}

type ProxyRequest struct {
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Service    string `json:"service"`
}

func (p *ProxyRequest) String() string {
	return fmt.Sprintf("%d:%s:%d", p.Port, p.Service, p.TargetPort)
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
