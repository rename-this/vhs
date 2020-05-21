package listener

import (
	"github.com/google/gopacket/pcap"
)

// NewListener creates a new listener.
func NewListener(addr *Addr, port uint16) (*Listener, error) {
	interfaces, err := addr.Interfaces()
	if err != nil {
		return nil, err
	}

	return &Listener{
		Port:       port,
		Interfaces: interfaces,
	}, nil
}

// Listener listens for TCP traffic on a
// given address and port.
type Listener struct {
	Addr       *Addr
	Port       uint16
	Interfaces []pcap.Interface
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

	for _, ii := range l.Interfaces {
		go func(i pcap.Interface) {
			errChan <- NewInterfaceError(i.Name, fn(i))
		}(ii)
	}

	for i := 0; i < len(l.Interfaces); i++ {
		allErrs.Append(<-errChan)
	}

	return allErrs
}

func (l *Listener) listen(i pcap.Interface) error {
	return nil
}
