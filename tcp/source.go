package tcp

import (
	"fmt"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/rename-this/vhs/capture"
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewSource creates a new TCP source.
func NewSource(_ session.Context) (flow.Source, error) {
	return &tcpSource{
		streams: make(chan flow.InputReader),
	}, nil
}

type tcpSource struct {
	streams chan flow.InputReader
}

func (s *tcpSource) Streams() <-chan flow.InputReader {
	return s.streams
}

func (s *tcpSource) Init(ctx session.Context) {
	s.read(ctx, capture.NewCapture, capture.NewListener)
}

type (
	newCaptureFn  func(addr string, response bool) (*capture.Capture, error)
	newListenerFn func(*capture.Capture) capture.Listener
)

func (s *tcpSource) read(ctx session.Context, newCapture newCaptureFn, newListener newListenerFn) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "tcp_source").
		Logger()

	ctx.Logger.Debug().Msg("read")

	cap, err := newCapture(ctx.FlowConfig.Addr, ctx.FlowConfig.CaptureResponse)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to initialize capture: %w", err)
		return
	}

	ctx.Logger.Debug().Interface("cap", cap).Msg("capture created")

	listener := newListener(cap)

	defer listener.Close()

	go listener.Listen(ctx)

	var (
		factory   = newStreamFactory(ctx, s.streams)
		pool      = tcpassembly.NewStreamPool(factory)
		assembler = tcpassembly.NewAssembler(pool)
		ticker    = time.Tick(ctx.FlowConfig.TCPTimeout)
		complete  = time.After(ctx.FlowConfig.SourceDuration)
		packets   = listener.Packets()
	)

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				if ctx.Config.DebugPackets {
					ctx.Logger.Debug().Msg("nil packet")
				}
				return
			}

			if packet.NetworkLayer() == nil ||
				packet.TransportLayer() == nil ||
				packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				if ctx.Config.DebugPackets {
					ctx.Logger.Debug().Str("p", packet.String()).Msg("wrong packet layers")
				}
				continue
			}

			var (
				tcp  = packet.TransportLayer().(*layers.TCP)
				flow = packet.NetworkLayer().NetworkFlow()
			)

			assembler.AssembleWithTimestamp(flow, tcp, time.Now())

		case <-ticker:
			ctx.Logger.Debug().Msg("flushing old streams")
			assembler.FlushOlderThan(time.Now().Add(-ctx.FlowConfig.TCPTimeout))

			factory.prune()

		case <-complete:
			factory.Close()
			return

		case <-ctx.StdContext.Done():
			return
		}
	}
}
