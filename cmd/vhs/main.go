package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gramLabs/vhs/capture"
)

func main() {
	var (
		addr = flag.String("addr", "0.0.0.0", "address to listen on")
		port = flag.Int("port", 0, "port to listen on")
	)

	flag.Parse()

	cap, err := capture.NewCapture(*addr, uint16(*port))
	if err != nil {
		log.Fatalf("failed to initialize capture: %v", err)
	}

	listener := capture.NewListener(cap)
	if err := listener.Listen(); err != nil {
		// TODO(andrewhare): Fix this API since not all interfaces
		// are guaranteed to work.
		// Maybe only print errors if all interfaces fail?
		log.Printf("failed to start listening: %v\n", err)
	}

	defer listener.Close()

	for p := range listener.Packets() {
		l := p.ApplicationLayer()
		if l != nil {
			fmt.Println(string(l.Payload()))
		}
	}
}
