package capture

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// NewListener creates a new listener.
func NewListener(cap *Capture) *Listener {
	return &Listener{
		Capture: cap,
		packets: make(chan gopacket.Packet),
	}
}

// Listener listens for TCP traffic on a
// given address and port.
type Listener struct {
	Capture *Capture

	packets chan gopacket.Packet

	handleMu sync.Mutex
	handles  []*pcap.Handle
}

// Listen starts listening.
func (l *Listener) Listen() error {
	return l.listenAll(l.listen)
}

// Packets retrieves a channel for all packets
// captured by this listener.
func (l *Listener) Packets() <-chan gopacket.Packet {
	return l.packets
}

type listenFn func(i pcap.Interface) error

func (l *Listener) listenAll(fn listenFn) error {
	var (
		allErrs = &Error{}
		errChan = make(chan *InterfaceError)
	)

	for _, ii := range l.Capture.Interfaces {
		go func(i pcap.Interface) {
			errChan <- NewInterfaceError(i.Name, fn(i))
		}(ii)
	}

	for i := 0; i < len(l.Capture.Interfaces); i++ {
		allErrs.Append(<-errChan)
	}

	if len(allErrs.InterfaceErrors) == 0 {
		return nil
	}

	return allErrs
}

func (l *Listener) listen(i pcap.Interface) error {
	handle, err := l.newActiveHandler(i.Name)
	if err != nil {
		return errors.Errorf("failed to get handle: %w", err)
	}

	if err := handle.SetBPFFilter(filter(l.Capture, i)); err != nil {
		return errors.Errorf("failed to set filter: %w", err)
	}

	l.handleMu.Lock()
	l.handles = append(l.handles, handle)
	l.handleMu.Unlock()

	go l.readPackets(handle, handle.LinkType())

	return nil
}

func (l *Listener) readPackets(dataSource gopacket.PacketDataSource, decoder gopacket.Decoder) {
	source := gopacket.NewPacketSource(dataSource, decoder)
	source.Lazy = true
	source.NoCopy = true

	for {
		p, err := source.NextPacket()
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			fmt.Printf("read packet failed: %v\n", err)
			continue
		}
		l.packets <- p
	}
}

func (l *Listener) newActiveHandler(name string) (*pcap.Handle, error) {
	inactive, err := pcap.NewInactiveHandle(name)
	if err != nil {
		return nil, errors.Errorf("failed to create inactive handle: %w", err)
	}

	defer inactive.CleanUp()

	inactive.SetPromisc(true)

	// TODO(andrewhare): configure these
	inactive.SetTimeout(2000 * time.Millisecond)
	inactive.SetImmediateMode(true)
	inactive.SetSnapLen(65536)

	handle, err := inactive.Activate()
	if err != nil {
		return nil, errors.Errorf("failed to activate handle: %w", err)
	}

	return handle, nil
}

// Close closes the listener and all open handles.
func (l *Listener) Close() {
	l.handleMu.Lock()
	defer l.handleMu.Unlock()

	for _, handle := range l.handles {
		handle.Close()
	}
}
