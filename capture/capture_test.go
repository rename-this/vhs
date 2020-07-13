package capture

import (
	"errors"
	"net"
	"testing"

	"github.com/google/gopacket/pcap"
	"gotest.tools/v3/assert"
)

func TestNewCapture(t *testing.T) {
	cases := []struct {
		desc        string
		addr        string
		port        string
		fn          getAllInterfacesFn
		capture     *Capture
		errContains string
	}{
		{
			desc: "success",
			addr: "1.1.1.1",
			port: "1111",
			fn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
					{
						Name: "111",
						Addresses: []pcap.InterfaceAddress{
							{IP: net.ParseIP("1.1.1.1")},
						},
					},
				}, nil
			},
			capture: &Capture{
				Host:       "1.1.1.1",
				Port:       "1111",
				DeviceType: CaptureIP,
				Interfaces: []pcap.Interface{
					{
						Name: "111",
						Addresses: []pcap.InterfaceAddress{
							{IP: net.ParseIP("1.1.1.1")},
						},
					},
				},
			},
		},
		{
			desc: "fail to get interfaces",
			addr: "1.1.1.1",
			port: "1111",
			fn: func() ([]pcap.Interface, error) {
				return nil, errors.New("111")
			},
			errContains: "111",
		},
		{
			desc: "fail to get capture type",
			addr: "1.1.1",
			port: "1111",
			fn: func() ([]pcap.Interface, error) {
				return nil, nil
			},
			errContains: "invalid address: 1.1.1",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			capture, err := newCapture(c.addr, c.port, c.fn)
			if err != nil {
				assert.ErrorContains(t, err, c.errContains)
				return
			}
			assert.DeepEqual(t, capture, c.capture)
		})
	}
}

func TestGeCaptureType(t *testing.T) {
	cases := []struct {
		desc        string
		device      string
		deviceType  Type
		errContains string
	}{
		{
			desc:       "empty",
			device:     "",
			deviceType: CaptureInvalid,
		},
		{
			desc:       "all zeros",
			device:     "0.0.0.0",
			deviceType: CaptureAll,
		},
		{
			desc:       "empty IPv6",
			device:     "::",
			deviceType: CaptureAll,
		},
		{
			desc:       "IPv4 loopback",
			device:     "127.0.0.1",
			deviceType: CaptureLoopback,
		},
		{
			desc:       "IPv6 loopback",
			device:     "::1",
			deviceType: CaptureLoopback,
		},
		{
			desc:       "regular",
			device:     "1.1.1.1",
			deviceType: CaptureIP,
		},
		{
			desc:        "invalid",
			device:      "1111",
			deviceType:  CaptureInvalid,
			errContains: "invalid address",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			deviceType, err := getCaptureType(c.device)
			if err != nil {
				assert.ErrorContains(t, err, c.errContains)
			}
			assert.Equal(t, c.deviceType, deviceType)
		})
	}
}

func TestSelectInterfaces(t *testing.T) {
	cases := []struct {
		desc               string
		value              string
		deviceType         Type
		interfaces         []pcap.Interface
		expectedInterfaces []pcap.Interface
	}{
		{
			desc:       "loopback",
			value:      "::1",
			deviceType: CaptureLoopback,
			interfaces: []pcap.Interface{
				{Name: "111"},
				{Name: "222"},
				{Name: "333"},
			},
			expectedInterfaces: []pcap.Interface{
				{Name: "111"},
				{Name: "222"},
				{Name: "333"},
			},
		},
		{
			desc:       "all with no addresses",
			value:      "::",
			deviceType: CaptureAll,
			interfaces: []pcap.Interface{
				{Name: "111"},
			},
			expectedInterfaces: nil,
		},
		{
			desc:       "all with addresses",
			value:      "::",
			deviceType: CaptureAll,
			interfaces: []pcap.Interface{
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
				{
					Name: "333",
				},
			},
			expectedInterfaces: []pcap.Interface{
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
		{
			desc:       "single address match name",
			value:      "1.1.1.1",
			deviceType: CaptureIP,
			interfaces: []pcap.Interface{
				{Name: "1.1.1.1"},
				{Name: "2.2.2.2"},
			},
			expectedInterfaces: []pcap.Interface{
				{Name: "1.1.1.1"},
			},
		},
		{
			desc:       "single address match IP",
			value:      "1.1.1.1",
			deviceType: CaptureIP,
			interfaces: []pcap.Interface{
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
			expectedInterfaces: []pcap.Interface{
				{
					Name: "111",
					Addresses: []pcap.InterfaceAddress{
						{IP: net.ParseIP("1.1.1.1")},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			interfaces := selectInterfaces(c.value, c.deviceType, c.interfaces)
			assert.DeepEqual(t, interfaces, c.expectedInterfaces)
		})
	}
}
