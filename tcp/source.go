package tcp

import (
	"time"

	"github.com/go-errors/errors"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/gramLabs/vhs/capture"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

// NewSource creates a new TCP source.
func NewSource(_ *session.Context) (flow.Source, error) {
	return &tcpSource{
		streams: make(chan ioutilx.ReadCloserID),
	}, nil
}

type tcpSource struct {
	streams chan ioutilx.ReadCloserID
}

func (s *tcpSource) Streams() <-chan ioutilx.ReadCloserID { return s.streams }

func (s *tcpSource) Init(ctx *session.Context) {
	s.read(ctx, capture.NewCapture, capture.NewListener)
}

type newCaptureFn func(addr string, response bool) (*capture.Capture, error)
type newListenereFn func(*capture.Capture) capture.Listener

func (s *tcpSource) read(ctx *session.Context, newCapture newCaptureFn, newListener newListenereFn) {
	cap, err := newCapture(ctx.Config.Addr, ctx.Config.CaptureResponse)
	if err != nil {
		ctx.Errors <- errors.Errorf("failed to initialize capture: %w", err)
		return
	}

	listener := newListener(cap)

	defer listener.Close()

	go listener.Listen(ctx)

	var (
		factory   = newStreamFactory(s.streams)
		pool      = tcpassembly.NewStreamPool(factory)
		assembler = tcpassembly.NewAssembler(pool)
		ticker    = time.Tick(ctx.Config.TCPTimeout)
		packets   = listener.Packets()
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
