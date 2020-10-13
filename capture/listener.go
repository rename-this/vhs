package capture

import (
	"io"
	"sync"

	"github.com/go-errors/errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/gramLabs/vhs/session"

	// See https://pkg.go.dev/github.com/google/gopacket?tab=doc#hdr-A_Final_Note
	_ "github.com/google/gopacket/layers"
)

// NewListener creates a new listener.
func NewListener(cap *Capture) Listener {
	return &listener{
		Capture: cap,
		packets: make(chan gopacket.Packet),
	}
}

// Listener listens for network traffic on a
// given address and port.
type Listener interface {
	Packets() <-chan gopacket.Packet
	Listen(session.Context)
	Close()
}

type listener struct {
	Capture *Capture

	packets chan gopacket.Packet

	handleMu sync.Mutex
	handles  []*pcap.Handle
}

// Packets retrieves a channel for all packets
// captured by this listener.
func (l *listener) Packets() <-chan gopacket.Packet {
	return l.packets
}

// Listen starts listening.
func (l *listener) Listen(ctx session.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "listener").
		Logger()

	for _, i := range l.Capture.Interfaces {
		if h, err := l.newHandle(ctx, i, (*pcap.InactiveHandle).Activate); err != nil {
			ctx.Errors <- err
		} else {
			go l.readPackets(ctx, h, h.LinkType())
		}
	}
}

type activateFn func(inactive *pcap.InactiveHandle) (*pcap.Handle, error)

func (l *listener) newHandle(ctx session.Context, i pcap.Interface, activate activateFn) (*pcap.Handle, error) {
	ctx.Logger = ctx.Logger.With().
		Interface("interface", i).
		Logger()

	ctx.Logger.Debug().Msg("creating new handle")

	inactive, err := l.newInactiveHandler(i.Name)
	if err != nil {
		ctx.Logger.Debug().Err(err).Msgf("failed to create inactive handle for %s", i.Name)
		return nil, nil
	}

	defer inactive.CleanUp()

	handle, err := activate(inactive)
	if err != nil {
		return nil, errors.Errorf("failed to activate %s: %w", i.Name, err)
	}

	ctx.Logger.Debug().Msg("handle activated")

	filter := newBPFFilter(l.Capture, i)
	ctx.Logger.Debug().Str("filter", filter).Msg("bpf filter created")

	if err := handle.SetBPFFilter(filter); err != nil {
		return nil, errors.Errorf("failed to set filter: %w", err)
	}

	l.handleMu.Lock()
	l.handles = append(l.handles, handle)
	l.handleMu.Unlock()

	return handle, nil
}

func (l *listener) readPackets(ctx session.Context, dataSource gopacket.PacketDataSource, decoder gopacket.Decoder) {
	source := gopacket.NewPacketSource(dataSource, decoder)
	source.Lazy = true
	source.NoCopy = true

	for {
		select {
		case <-ctx.StdContext.Done():
			return
		default:
			p, err := source.NextPacket()
			if errors.Is(err, io.EOF) {
				continue
			}
			if err != nil {
				if ctx.Config.DebugPackets {
					ctx.Logger.Debug().Err(err).Msg("read packet failed")
				}
				continue
			}
			if ctx.Config.DebugPackets {
				ctx.Logger.Debug().Str("p", p.String()).Msg("packet")
			}
			l.packets <- p
		}
	}
}

func (l *listener) newInactiveHandler(name string) (*pcap.InactiveHandle, error) {
	inactive, err := pcap.NewInactiveHandle(name)
	if err != nil {
		return nil, errors.Errorf("failed to create inactive handle: %w", err)
	}

	inactive.SetPromisc(true)
	inactive.SetTimeout(pcap.BlockForever)
	inactive.SetImmediateMode(true)
	inactive.SetSnapLen(65536)

	return inactive, nil
}

// Close closes the listener and all open handles.
func (l *listener) Close() {
	l.handleMu.Lock()
	defer l.handleMu.Unlock()

	for _, handle := range l.handles {
		handle.Close()
	}
}
