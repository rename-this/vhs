package capture

import (
	"net"

	"github.com/go-errors/errors"
	"github.com/google/gopacket/pcap"
)

const (
	// DefaultAddr is the default capture address.
	DefaultAddr = "0.0.0.0:80"
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
	Host       string
	Port       string
	DeviceType Type
	Interfaces []pcap.Interface

	Response bool
}

type getAllInterfacesFn func() ([]pcap.Interface, error)

// NewCapture creates a new capture.
func NewCapture(addr string, response bool) (*Capture, error) {
	return newCapture(addr, response, pcap.FindAllDevs)
}

func newCapture(addr string, response bool, fn getAllInterfacesFn) (*Capture, error) {
	interfaces, err := fn()
	if err != nil {
		return nil, errors.Errorf("failed to find interfaces: %w", err)
	}

	host, port := splitHostPort(addr)
	deviceType, err := getCaptureType(host)
	if err != nil {
		return nil, errors.Errorf("failed to get device type: %w", err)
	}

	return &Capture{
		Host:       host,
		Port:       port,
		DeviceType: deviceType,
		Response:   response,
		Interfaces: selectInterfaces(host, deviceType, interfaces),
	}, nil
}

func selectInterfaces(host string, deviceType Type, interfaces []pcap.Interface) []pcap.Interface {
	if deviceType == CaptureLoopback {
		return interfaces
	}

	var filtered []pcap.Interface
	for _, i := range interfaces {
		if deviceType == CaptureAll && len(i.Addresses) > 0 {
			filtered = append(filtered, i)
			continue
		}

		if i.Name == host {
			return []pcap.Interface{i}
		}

		for _, address := range i.Addresses {
			if address.IP.String() == host {
				return []pcap.Interface{i}
			}
		}
	}

	return filtered
}

func getCaptureType(host string) (Type, error) {
	ip := net.ParseIP(host)
	if ip.IsLoopback() {
		return CaptureLoopback, nil
	}
	if ip.IsUnspecified() {
		return CaptureAll, nil
	}
	if ip != nil {
		return CaptureIP, nil
	}

	return CaptureInvalid, errors.Errorf("invalid address: %s", host)
}

func splitHostPort(addr string) (string, string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// TODO(andrewhare): Log that we're using the default addr.
		host, port, _ = net.SplitHostPort(DefaultAddr)
	}
	return host, port
}
