package capture

import (
	"fmt"
	"sync"

	"github.com/google/gopacket/pcap"
)

// NewListener creates a new listener.
func NewListener(addr *Capture, port uint16) *Listener {
	return &Listener{
		Capture: addr,
		Port:    port,
	}
}

// Listener listens for TCP traffic on a
// given address and port.
type Listener struct {
	Capture *Capture
	Port    uint16

	handleMu sync.Mutex
	handles  []*pcap.Handle
}

// Listen starts listening.
func (l *Listener) Listen() error {
	return l.listenAll(l.listen)
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

	return allErrs
}

func (l *Listener) listen(i pcap.Interface) error {
	handle, err := l.newActiveHandler(i.Name)
	if err != nil {
		return fmt.Errorf("failed to get handle: %w", err)
	}

	defer handle.Close()

	l.handleMu.Lock()
	l.handles = append(l.handles, handle)
	l.handleMu.Unlock()

	return nil
}

func (l *Listener) newActiveHandler(name string) (*pcap.Handle, error) {
	inactive, err := pcap.NewInactiveHandle(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create inactive handle: %w", err)
	}

	defer inactive.CleanUp()

	//
	// TODO(andrewhare): Configure the handle.
	//

	handle, err := inactive.Activate()
	if err != nil {
		return nil, fmt.Errorf("failed to activate handle: %w", err)
	}

	return handle, nil
}
