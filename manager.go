package main

import (
	"errors"
	"log"
	"sync"
)

var proxyManager Manager

func init() {
	proxyManager = newProxyManager()
}

func newProxyManager() Manager {
	return Manager{
		Mutex:         sync.RWMutex{},
		ProxyMap:      make(map[string]*UDPProxy),
		PortsOccupied: make(map[int]bool),
	}
}

func (m *Manager) Add(pr ProxyRecord) (bool, error) {
	m.Mutex.RLock()
	// check if already exists on same port
	if _, ok := m.PortsOccupied[pr.Port]; ok {
		// check if proxy is already present
		n := pr.String()
		if _, ok := m.ProxyMap[n]; ok {
			return true, nil
		} else {
			return false, errors.New("port already occupied")
		}
	}
	m.Mutex.RUnlock()
	// create a new proxy
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	p, err := NewUDPProxy(pr.Port, pr.TargetPort, pr.Service)
	if err != nil {
		return false, errors.New("failed to create proxy")
	}
	// add the proxy to the map
	m.ProxyMap[pr.String()] = p
	m.PortsOccupied[pr.Port] = true
	// store record in file
	ll := UDPProxyToRecordList(m.ProxyMap)
	err = StoreRecordsInFile(ll)
	if err != nil {
		delete(m.ProxyMap, pr.String())
		return false, errors.New("failed to update records")
	}
	// start the proxy
	go p.Run()
	// return success
	return true, nil
}

func (m *Manager) Remove(pr ProxyRecord) (bool, error) {
	// handle panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in remove:", r)
			return
		}
	}()
	m.Mutex.RLock()
	// check if the proxy exists
	if _, ok := m.ProxyMap[pr.String()]; !ok {
		m.Mutex.RUnlock()
		return false, errors.New("proxy does not exist")
	}
	m.Mutex.RUnlock()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	// close the proxy
	m.ProxyMap[pr.String()].Close()
	// remove the proxy from the map
	delete(m.ProxyMap, pr.String())
	// remove the port from the occupied ports
	delete(m.PortsOccupied, pr.Port)
	// store record in file
	ll := UDPProxyToRecordList(m.ProxyMap)
	err := StoreRecordsInFile(ll)
	if err != nil {
		return false, errors.New("failed to update records")
	}
	// return success
	return true, nil
}

func (m *Manager) Exist(pr ProxyRecord) bool {
	m.Mutex.RLock()
	// check if the proxy exists
	_, ok := m.ProxyMap[pr.String()]
	m.Mutex.RUnlock()
	return ok
}

func (m *Manager) PortOccupied(port int) bool {
	m.Mutex.RLock()
	// check if the port is occupied
	_, ok := m.PortsOccupied[port]
	m.Mutex.RUnlock()
	return ok
}

func (m *Manager) List() []string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	var list []string = make([]string, 0, len(m.ProxyMap))
	for k := range m.ProxyMap {
		list = append(list, k)
	}
	return list
}

func (m *Manager) Close() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	for _, p := range m.ProxyMap {
		p.Close()
	}
}
