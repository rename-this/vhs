package capture

import (
	"fmt"
	"strings"

	"github.com/google/gopacket/pcap"
)

// newBPFFilter creates a BPF newBPFFilter based on a capture configuration
// and a given interface.
func newBPFFilter(capture *Capture, iface pcap.Interface) string {
	var addrs []string

	switch capture.DeviceType {
	case CaptureLoopback:
		for _, i := range capture.Interfaces {
			for _, a := range i.Addresses {
				addrs = append(addrs, fmt.Sprintf("host %s", a.IP))
			}
		}
	default:
		for _, a := range iface.Addresses {
			addrs = append(addrs, fmt.Sprintf("host %s", a.IP.String()))
		}
	}

	hosts := strings.Join(addrs, " or ")

	switch l := len(addrs); {
	case l > 1:
		// Wrap the hosts in ( )
		hosts = fmt.Sprintf("(%s)", hosts)
	case l == 1:
		// No special formatting needed
	default:
		// No hosts/addrs means no filters
		return ""
	}

	portExpression := "dst port"
	if capture.Response {
		portExpression = "port"
	}

	return fmt.Sprintf("tcp %s %s and %s", portExpression, capture.Port, hosts)
}
