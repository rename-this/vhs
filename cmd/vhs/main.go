package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/http"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/sink"
)

func main() {
	var (
		addr      = flag.String("addr", "0.0.0.0", "address to listen on")
		port      = flag.Int("port", 0, "port to listen on")
		mwarePath = flag.String("middleware", "", "path to a middleware executable")
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

	mware, err := middleware.New(context.TODO(), *mwarePath, os.Stderr)
	if err != nil {
		log.Fatalf("failed to initialize middleware: %v", err)
	}

	defer mware.Close()

	go func() {
		if err := mware.Start(); err != nil {
			log.Fatalf("failed to start middleware: %v", err)
		}
	}()

	var (
		sinks = []sink.Sink{
			sink.NewStdout(),
		}
		factory = &http.StreamFactory{
			Middleware: mware,
			Sinks:      sinks,
		}
		pool      = tcpassembly.NewStreamPool(factory)
		assembler = tcpassembly.NewAssembler(pool)
		packets   = listener.Packets()
		ticker    = time.Tick(time.Minute)
	)

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				continue
			}

			var (
				tcp  = packet.TransportLayer().(*layers.TCP)
				flow = packet.NetworkLayer().NetworkFlow()
			)

			assembler.AssembleWithTimestamp(flow, tcp, time.Now())

		case <-ticker:
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
