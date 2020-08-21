package tcp

import (
	"io"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/source"
)

// NewSource creates a new TCP source.
func NewSource(_ *session.Context) (source.Source, error) {
	return &tcpSource{
		streams: make(chan io.ReadCloser),
	}, nil
}

type tcpSource struct {
	streams chan io.ReadCloser
}

func (s *tcpSource) Streams() <-chan io.ReadCloser { return s.streams }

func (s *tcpSource) Init(ctx *session.Context) {
	cap, err := capture.NewCapture(ctx.Config.Addr, ctx.Config.CaptureResponse)
	if err != nil {
		ctx.Errors <- errors.Errorf("failed to initialize capture: %w", err)
		return
	}

	listener := capture.NewListener(cap)
	defer listener.Close()

	go listener.Listen(ctx)

	var (
		factory   = newStreamFactory(s.streams)
		pool      = tcpassembly.NewStreamPool(factory)
		assembler = tcpassembly.NewAssembler(pool)
		packets   = listener.Packets()
		ticker    = time.Tick(ctx.Config.TCPTimeout)
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
			assembler.FlushOlderThan(time.Now().Add(-ctx.Config.TCPTimeout))
			factory.prune()

		case <-ctx.StdContext.Done():
			return
		}
	}
}
