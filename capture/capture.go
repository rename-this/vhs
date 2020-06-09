package capture

import (
	"fmt"
	"net"

	"github.com/google/gopacket/pcap"
)

// Type represents the type of device.
type Type int

const (
	// CaptureInvalid is an invalid device type.
	CaptureInvalid Type = iota
	// CaptureIP is a standard IP address.
	CaptureIP
	// CaptureLoopback represents the loopback interface (e.g. 127.0.0.1).
	CaptureLoopback
	// CaptureAll represents all interfaces (e.g. 0.0.0.0).
	CaptureAll
)

// Capture represents an intent to capture traffic.
type Capture struct {
	Addr       string
	Port       uint16
	DeviceType Type
	Interfaces []pcap.Interface

	Response bool
}

type getAllInterfacesFn func() ([]pcap.Interface, error)

// NewCapture creates a new capture.
func NewCapture(addr string, port uint16) (*Capture, error) {
	return newCapture(addr, port, pcap.FindAllDevs)
}

func newCapture(addr string, port uint16, fn getAllInterfacesFn) (*Capture, error) {
	interfaces, err := fn()
	if err != nil {
		return nil, fmt.Errorf("failed to find interfaces: %w", err)
	}

	deviceType, err := getCaptureType(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get device type: %w", err)
	}

	return &Capture{
		Addr:       addr,
		Port:       port,
		DeviceType: deviceType,
		Interfaces: selectInterfaces(addr, deviceType, interfaces),
	}, nil
}

func selectInterfaces(addr string, deviceType Type, interfaces []pcap.Interface) []pcap.Interface {
	if deviceType == CaptureLoopback {
		return interfaces
	}

	var filtered []pcap.Interface
	for _, i := range interfaces {
		if deviceType == CaptureAll && len(i.Addresses) > 0 {
			filtered = append(filtered, i)
			continue
		}

		if i.Name == addr {
			return []pcap.Interface{i}
		}

		for _, address := range i.Addresses {
			if address.IP.String() == addr {
				return []pcap.Interface{i}
			}
		}
	}

	return filtered
}

func getCaptureType(addr string) (Type, error) {
	switch addr {
	case "", "0.0.0.0", "::":
		return CaptureAll, nil
	case "127.0.0.1", "::1":
		return CaptureLoopback, nil
	default:
		if ip := net.ParseIP(addr); ip != nil {
			return CaptureIP, nil
		}
	}
	return CaptureInvalid, fmt.Errorf("invalid address: %s", addr)
}
