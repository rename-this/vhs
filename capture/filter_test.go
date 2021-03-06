package capture

import (
	"net"
	"testing"

	"github.com/google/gopacket/pcap"
	"gotest.tools/v3/assert"
)

func TestNewBPFFilter(t *testing.T) {
	cases := []struct {
		desc           string
		capture        *Capture
		iface          pcap.Interface
		filter         string
		responseFilter string
	}{
		{
			desc: "loopback no addrs",
			capture: &Capture{
				DeviceType: CaptureLoopback,
			},
			filter:         "",
			responseFilter: "",
		},
		{
			desc: "loopback with single addr",
			capture: &Capture{
				Port:       "1111",
				DeviceType: CaptureLoopback,
				Interfaces: []pcap.Interface{
					{
						Name: "111",
						Addresses: []pcap.InterfaceAddress{
							{IP: net.ParseIP("1.1.1.1")},
						},
					},
				},
			},
			filter:         "tcp dst port 1111 and host 1.1.1.1",
			responseFilter: "tcp port 1111 and host 1.1.1.1",
		},
		{
			desc: "loopback with multiple addrs",
			capture: &Capture{
				Port:       "1111",
				DeviceType: CaptureLoopback,
				Interfaces: []pcap.Interface{
					{
						Name: "111",
						Addresses: []pcap.InterfaceAddress{
							{IP: net.ParseIP("1.1.1.1")},
						},
					},
					{
						Name: "222",
						Addresses: []pcap.InterfaceAddress{
							{IP: net.ParseIP("2.2.2.2")},
						},
					},
				},
			},
			filter:         "tcp dst port 1111 and (host 1.1.1.1 or host 2.2.2.2)",
			responseFilter: "tcp port 1111 and (host 1.1.1.1 or host 2.2.2.2)",
		},
		{
			desc: "interface with single addr",
			capture: &Capture{
				Port:       "1111",
				DeviceType: CaptureIP,
			},
			iface: pcap.Interface{
				Name: "111",
				Addresses: []pcap.InterfaceAddress{
					{IP: net.ParseIP("1.1.1.1")},
				},
			},
			filter:         "tcp dst port 1111 and host 1.1.1.1",
			responseFilter: "tcp port 1111 and host 1.1.1.1",
		},
		{
			desc: "interface with multiple addrs",
			capture: &Capture{
				Port:       "1111",
				DeviceType: CaptureIP,
			},
			iface: pcap.Interface{
				Name: "111",
				Addresses: []pcap.InterfaceAddress{
					{IP: net.ParseIP("1.1.1.1")},
					{IP: net.ParseIP("2.2.2.2")},
				},
			},
			filter:         "tcp dst port 1111 and (host 1.1.1.1 or host 2.2.2.2)",
			responseFilter: "tcp port 1111 and (host 1.1.1.1 or host 2.2.2.2)",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.filter, newBPFFilter(c.capture, c.iface))
			c.capture.Response = true
			assert.Equal(t, c.responseFilter, newBPFFilter(c.capture, c.iface))
		})
	}
}
