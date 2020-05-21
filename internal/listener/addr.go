package listener

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

// AddrType represents the type of address.
type AddrType int

const (
	// AddrRegular is a standard IP address.
	AddrRegular AddrType = iota
	// AddrLoopback represents the loopback interface (e.g. 127.0.0.1).
	AddrLoopback
	// AddrAll represents all interfaces (e.g. 0.0.0.0).
	AddrAll
)

// Addr is a network address
type Addr struct {
	Value string
	Type  AddrType

	getInterfacesFn func() ([]pcap.Interface, error)
}

// NewAddr creates a new address.
func NewAddr(addr string) *Addr {
	return &Addr{
		Value:           addr,
		Type:            getAddrType(addr),
		getInterfacesFn: pcap.FindAllDevs,
	}
}

// Interfaces gets all interfaces for the current address.
func (a *Addr) Interfaces() ([]pcap.Interface, error) {
	interfaces, err := a.getInterfacesFn()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	if a.Type == AddrLoopback {
		return interfaces, nil
	}

	var filtered []pcap.Interface
	for _, i := range interfaces {
		if a.Type == AddrAll && len(i.Addresses) > 0 {
			filtered = append(filtered, i)
			continue
		}

		if i.Name == a.Value {
			return []pcap.Interface{i}, nil
		}

		for _, address := range i.Addresses {
			if address.IP.String() == a.Value {
				return []pcap.Interface{i}, nil
			}
		}
	}

	return filtered, nil
}

func getAddrType(addr string) AddrType {
	switch addr {
	case "", "0.0.0.0", "::":
		return AddrAll
	case "127.0.0.1", "::1":
		return AddrLoopback
	default:
		return AddrRegular
	}
}
