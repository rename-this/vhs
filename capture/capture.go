package capture

import (
	"fmt"
	"strings"

	"github.com/google/gopacket/pcap"
)

// DeviceType represents the type of device.
type DeviceType int

const (
	// DeviceRegular is a standard IP address.
	DeviceRegular DeviceType = iota
	// DeviceLoopback represents the loopback interface (e.g. 127.0.0.1).
	DeviceLoopback
	// DeviceAll represents all interfaces (e.g. 0.0.0.0).
	DeviceAll
)

// Capture represents an intent to capture traffic.
type Capture struct {
	Device     string
	Port       uint16
	DeviceType DeviceType
	Interfaces []pcap.Interface

	CaptureResponse bool
}

// NewCapture creates a new capture.
func NewCapture(device string, port uint16) (*Capture, error) {
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("failed to find interfaces: %w", err)
	}

	var (
		deviceType = getDeviceType(device)
		selected   = selectInterfaces(device, deviceType, interfaces)
	)

	return &Capture{
		Device:     device,
		Port:       port,
		DeviceType: deviceType,
		Interfaces: selected,
	}, nil
}

// Filter creates a BPF string for the current addr.
func (a *Capture) Filter() string {
	var (
		src string
		dst string
	)

	switch a.DeviceType {
	case DeviceLoopback:
		var addrs []string
		for _, dc := range a.Interfaces {
			for _, addr := range dc.Addresses {
				addrs = append(addrs, fmt.Sprintf("(dst host %s and src host %s)", addr.IP, addr.IP))
			}
		}
		dst = strings.Join(addrs, " or ")
		src = dst
	default:
	}

	_ = src

	return fmt.Sprintf("tcp dst port %d and (%s)", a.Port, dst)
}

func selectInterfaces(name string, deviceType DeviceType, interfaces []pcap.Interface) []pcap.Interface {
	if deviceType == DeviceLoopback {
		return interfaces
	}

	var filtered []pcap.Interface
	for _, i := range interfaces {
		if deviceType == DeviceAll && len(i.Addresses) > 0 {
			filtered = append(filtered, i)
			continue
		}

		if i.Name == name {
			return []pcap.Interface{i}
		}

		for _, address := range i.Addresses {
			if address.IP.String() == name {
				return []pcap.Interface{i}
			}
		}
	}

	return filtered
}

func getDeviceType(device string) DeviceType {
	switch device {
	case "", "0.0.0.0", "::":
		return DeviceAll
	case "127.0.0.1", "::1":
		return DeviceLoopback
	default:
		return DeviceRegular
	}
}
