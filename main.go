package main

import (
	"log"
)

func main() {
	p, err := NewUDPProxy(1212, 51820, "ip-13-200-80-4.swiftwave.xyz")
	if err != nil {
		log.Println(err)
	}
	// go func ()  {
	// 	// wait for 30 seconds
	// 	time.Sleep(30 * time.Second)
	// 	p.Close()
	// }()
	p.Run()

}
