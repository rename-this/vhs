package main

import (
	"context"
	"log"
	_http "net/http"
	"os"
	"strings"
	"time"

	"github.com/gramLabs/vhs/http"

	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/output"
	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/sink"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/tcp"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	tcpTimeout  = 5 * time.Minute
	httpTimeout = 30 * time.Second
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record network traffic",
	Run:   record,
}

func record(cmd *cobra.Command, args []string) {

	//Start Prometheus endpoint if requested by the user.
	if promAddr != "" {
		_http.Handle("/metrics", promhttp.Handler())
		go func() {
			log.Fatal(_http.ListenAndServe(promAddr, nil))
		}()
	}

	// TOOD(andrehare): Use this context to coordinate
	// all the pieces of the recording.
	ctx, cancel := context.WithCancel(context.Background())
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

	pipes := output.Pipes{
		output.NewPipe(format.NewJSON(), os.Stdout),
	}

	if gcsProjectID != "" {
		var (
			har      = http.NewHAR(30 * time.Second)
			gcs, err = sink.NewGCS(ctx, sink.GCSConfig{
				Session:    sess,
				ProjectID:  gcsProjectID,
				BucketName: gcsBucketName,
			})
		)
		if err != nil {
			log.Fatalf("failed to initialize GCS sink: %v", err)
		}
		pipes = append(pipes, output.NewPipe(har, gcs))
	}

	// Add the metrics pipe if the user has enabled Prometheus metrics.
	if promAddr != "" {
		pipes = append(pipes, http.NewMetricsPipe(httpTimeout))
	}

	for _, p := range pipes {
		go p.Init(ctx)
		defer func(s sink.Sink) {
			if err := s.Close(); err != nil {
				log.Printf("failed to close sink: %v\n", err)
			}
		}(p.Sink)
	}

	switch strings.ToLower(protocol) {
	case "http":
		factory := newStreamFactoryHTTP(ctx, sess, pipes)
		if factory.Middleware != nil {
			defer factory.Middleware.Close()
		}
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
