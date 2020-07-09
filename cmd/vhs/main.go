package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/http"
)

func main() {
	var (
		addr      = flag.String("addr", "0.0.0.0", "address to listen on")
		port      = flag.Int("port", 0, "port to listen on")
		mwarePath = flag.String("middleware", "", "path to a middleware executable")
	)

	flag.Parse()

	middleware, err := http.NewMiddleware(context.TODO(), *mwarePath)
	if err != nil {
		log.Fatalf("failed to initialize middleware: %v", err)
	}
	middleware.Stderr = os.Stderr

	defer middleware.Close()

	go func() {
		if err := middleware.Start(); err != nil {
			log.Fatalf("failed to start middleware: %v", err)
		}
	}()

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
