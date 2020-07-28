package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/http"
	"github.com/gramLabs/vhs/output"
	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/sink"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/tcp"
	"github.com/spf13/cobra"
)

const (
	tcpTimeout = 5 * time.Minute
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record network traffic",
	Run:   record,
}

func record(cmd *cobra.Command, args []string) {
	// TOOD(andrehare): Use this context to coordinate
	// all the pieces of the recording.
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	sess := session.New()

	cap, err := capture.NewCapture(address)
	if err != nil {
		log.Fatalf("failed to initialize capture: %v", err)
	}

	cap.Response = captureResponse

	listener := capture.NewListener(cap)
	if err := listener.Listen(); err != nil {
		// TODO(andrewhare): Fix this API since not all interfaces
		// are guaranteed to work.
		// Maybe only print errors if all interfaces fail?
		log.Printf("failed to start listening: %v\n", err)
	}

	defer listener.Close()

	pipes := pipes()
	for _, p := range pipes {
		p.Init(ctx)
	}

	switch strings.ToLower(protocol) {
	case "http":
		factory := newStreamFactoryHTTP(ctx, sess, pipes)
		defer factory.Middleware.Close()
		recordTCP(listener, factory)
	default:
		log.Fatal("invalid protocol")
	}
}

func recordTCP(listener *capture.Listener, factory tcp.BidirectionalStreamFactory) {
	var (
		pool      = tcpassembly.NewStreamPool(factory)
		assembler = tcpassembly.NewAssembler(pool)
		packets   = listener.Packets()
		ticker    = time.Tick(tcpTimeout)
	)

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}

			if packet.NetworkLayer() == nil ||
				packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				continue
			}

			var (
				tcp  = packet.TransportLayer().(*layers.TCP)
				flow = packet.NetworkLayer().NetworkFlow()
			)

			assembler.AssembleWithTimestamp(flow, tcp, time.Now())

		case <-ticker:
			assembler.FlushOlderThan(time.Now().Add(-tcpTimeout))
			factory.Prune(tcpTimeout)
		}
	}
}

func newStreamFactoryHTTP(ctx context.Context, sess *session.Session, pipes []*output.Pipe) *http.StreamFactory {
	var (
		m   *http.Middleware
		err error
	)

	if middleware != "" {
		m, err = http.NewMiddleware(ctx, middleware, os.Stderr)
		if err != nil {
			log.Fatalf("failed to initialize middleware: %v", err)
		}

		go func() {
			if err := m.Start(); err != nil {
				log.Fatalf("failed to start middleware: %v", err)
			}
		}()
	}

	return http.NewStreamFactory(m, sess, pipes)
}

func pipes() []*output.Pipe {
	return []*output.Pipe{
		output.NewPipe(format.NewJSON(), &sink.Stdout{}),
	}
}
