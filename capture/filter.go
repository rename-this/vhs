package capture

import (
	"fmt"
	"strings"

	"github.com/google/gopacket/pcap"
)

// newBPFFilter creates a BPF newBPFFilter based on a capture configuration
// and a given interface.
func newBPFFilter(capture *Capture, iface pcap.Interface) string {
	var (
		dst string
		src string
	)

	switch capture.DeviceType {
	case CaptureLoopback:
		var addrs []string
		for _, i := range capture.Interfaces {
			for _, a := range i.Addresses {
				addrs = append(addrs, fmt.Sprintf("(dst host %s and src host %s)", a.IP, a.IP))
			}
		}
		src = strings.Join(addrs, " or ")
		dst = src
	default:
		for i, a := range iface.Addresses {
			src += "dst host " + a.IP.String()
			dst += "src host " + a.IP.String()
			if i != len(iface.Addresses)-1 {
				dst += " or "
				src += " or "
			}
		}
	}

	if src == "" {
		return ""
	}

	if capture.Response {
		return fmt.Sprintf("(tcp dst port %s and (%s)) or (tcp src port %s and (%s))", capture.Port, dst, capture.Port, src)
	}

	return fmt.Sprintf("tcp dst port %s and (%s)", capture.Port, src)
}
