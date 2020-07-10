package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/http"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/sink"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(recordCmd)
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record network traffic",
	Run:   record,
}

func record(cmd *cobra.Command, args []string) {
	cap, err := capture.NewCapture(address)
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

	var mware *middleware.Middleware
	if middlewarePath != "" {
		mware, err = middleware.New(context.TODO(), middlewarePath, os.Stderr)
		if err != nil {
			log.Fatalf("failed to initialize middleware: %v", err)
		}

		defer mware.Close()

		go func() {
			if err := mware.Start(); err != nil {
				log.Fatalf("failed to start middleware: %v", err)
			}
		}()
	}

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
