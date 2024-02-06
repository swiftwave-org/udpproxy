package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

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

	server.Start(":8080")
	// wait lifetime
	select {}
}
