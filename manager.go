package main

import (
	"errors"
	"sync"
)

func newManager() Manager {
	return Manager{
		Mutex:    sync.RWMutex{},
		ProxyMap: make(map[string]*UDPProxy),
	}
}

func (m *Manager) addProxy(pr ProxyRequest) (bool, error) {
	m.Mutex.RLock()
	// check if the proxy already exists
	if _, ok := m.ProxyMap[pr.String()]; ok {
		m.Mutex.RUnlock()
		return false, errors.New("proxy already exists")
	}
	m.Mutex.RUnlock()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	// create a new proxy
	// p, err := NewUDPProxy(&net.UDPAddr{
	// 	IP:   net.ParseIP("0.0.0.0"),
	// 	Port: pr.Port,
	// 	Zone: "",
	// }, &net.UDPAddr{
	// 	IP: net.ParseIP(""),
	// })

	return true, nil
}
