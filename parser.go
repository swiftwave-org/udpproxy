package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var storedFilePath = "/var/lib/udpproxy/records"

func init() {
	// try to read records path from environment
	storedFilePathEnv := os.Getenv("RECORDS_PATH")
	if storedFilePathEnv != "" {
		storedFilePath = storedFilePathEnv
	}
	// create directory if not exists
	directory := filepath.Dir(storedFilePath)
	// create directory if not exists
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		log.Println("failed to create directory")
		log.Println(err)
		os.Exit(1)
	}
}

func StoreRecordsInFile(proxyRequestList []ProxyRecord) error {
	// prepare string to write in file
	var data string
	for _, proxyRequest := range proxyRequestList {
		data += fmt.Sprintf("%d %s:%d\n", proxyRequest.Port, proxyRequest.Service, proxyRequest.TargetPort)
	}
	// write in file
	err := writeToFile(storedFilePath, data)
	if err != nil {
		return err
	}
	return nil
}

func ReadRecordsFromFile() ([]ProxyRecord, error) {
	// read from file
	data, err := readFromFile(storedFilePath)
	if err != nil {
		return nil, err
	}
	// parse data
	splitDataInLines := strings.Split(data, "\n")
	var proxyRequestList []ProxyRecord = make([]ProxyRecord, 0)
	for _, line := range splitDataInLines {
		proxyRequest, err := parseLine(line)
		if err != nil {
			log.Println("failed to parse line")
			log.Println(err)
		} else {
			proxyRequestList = append(proxyRequestList, proxyRequest)
		}
	}
	if err != nil {
		return nil, err
	}
	return proxyRequestList, nil

}

func writeToFile(path string, data string) error {
	// write in file
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		log.Println("failed to write in file")
		log.Println(err)
		return errors.New("failed to write in file")
	}
	return nil
}

func readFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// check if path does not exist
		if os.IsNotExist(err) {
			return "", nil
		}
		log.Println("failed to read from file")
		log.Println(err)
		return "", errors.New("failed to read from file")
	}
	d := string(data)
	d = strings.TrimSpace(d)
	return d, nil
}

func parseLine(line string) (ProxyRecord, error) {
	var proxyRequest ProxyRecord = ProxyRecord{}
	line = strings.TrimSpace(line)
	// skip empty lines
	if line == "" {
		return proxyRequest, errors.New("empty line")
	}
	// split line
	splitLine := strings.Split(line, " ")
	if len(splitLine) != 2 {
		return proxyRequest, errors.New("invalid line")
	}
	// parse port
	port, err := strconv.Atoi(splitLine[0])
	if err != nil {
		return proxyRequest, errors.New("invalid port")
	}
	// parse service and target port
	splitServiceAndTargetPort := strings.Split(splitLine[1], ":")
	if len(splitServiceAndTargetPort) != 2 {
		return proxyRequest, errors.New("invalid service and target port")
	}
	service := splitServiceAndTargetPort[0]
	targetPort, err := strconv.Atoi(splitServiceAndTargetPort[1])
	if err != nil {
		return proxyRequest, errors.New("invalid target port")
	}
	// set values
	proxyRequest.Port = port
	proxyRequest.Service = service
	proxyRequest.TargetPort = targetPort
	return proxyRequest, nil
}
