package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var socketPath = "/etc/udplb/api.sock"

func init() {
	socketPathEnv := os.Getenv("SOCKET_PATH")
	if socketPathEnv != "" {
		socketPath = socketPathEnv
	}
	// create dir
	err := os.MkdirAll(filepath.Dir(socketPath), 0755)
	if err != nil {
		log.Println("Failed to create directory")
		log.Println(err)
		os.Exit(1)
	}
}

func main() {
	server := newAPIServer()
	// load records from file
	records, err := ReadRecordsFromFile()
	if err != nil {
		log.Println("Failed to read records from file")
		log.Println(err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	// on SIGTERM, or SIGKILL, stop the server
	go func() {
		sig := <-sigs
		go func() {
			for sig := range sigs {
				log.Printf("Already received %s, ignoring", sig)
			}
		}()
		log.Printf("Received %s, stopping server", sig)
		_ = server.Close()
		log.Println("Server stopped")
		log.Println("Stopping all proxies")
		proxyManager.Close()
		log.Println("Proxies stopped")
		os.Exit(0)
	}()

	go func() {
		// add records to manager
		for _, record := range records {
			_, err := proxyManager.Add(record)
			if err != nil {
				log.Println("Failed to add record")
				log.Println(err)
			}
		}
	}()

	// start server
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	server.Listener = listener
	httpServer := new(http.Server)
	if err := server.StartServer(httpServer); err != nil {
		log.Fatal(err)
	}
	// wait lifetime
	select {}
}
