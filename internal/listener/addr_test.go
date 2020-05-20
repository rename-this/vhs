package listener

import (
	"errors"
	"net"
	"testing"

	"github.com/google/gopacket/pcap"
	"gotest.tools/v3/assert"
)

func TestGetAddrType(t *testing.T) {
	cases := []struct {
		desc     string
		value    string
		addrType AddrType
	}{
		{
			desc:     "empty",
			value:    "",
			addrType: AddrAll,
		},
		{
			desc:     "all zeros",
			value:    "0.0.0.0",
			addrType: AddrAll,
		},
		{
			desc:     "empty IPv6",
			value:    "::",
			addrType: AddrAll,
		},
		{
			desc:     "IPv4 loopback",
			value:    "127.0.0.1",
			addrType: AddrLoopback,
		},
		{
			desc:     "IPv6 loopback",
			value:    "::1",
			addrType: AddrLoopback,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			addrType := getAddrType(c.value)
			assert.Equal(t, c.addrType, addrType)
		})
	}
}

func TestInterfaces(t *testing.T) {
	cases := []struct {
		desc            string
		value           string
		getInterfacesFn func() ([]pcap.Interface, error)
		errContains     string
		interfaces      []pcap.Interface
	}{
		{
			desc: "get devices err",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return nil, errors.New("111")
			},
			errContains: "111",
		},
		{
			desc:  "loopback",
			value: "::1",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
					{Name: "111"},
					{Name: "222"},
					{Name: "333"},
				}, nil
			},
			interfaces: []pcap.Interface{
				{Name: "111"},
				{Name: "222"},
				{Name: "333"},
			},
		},
		{
			desc:  "all with no addresses",
			value: "::",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
					{Name: "111"},
				}, nil
			},
		},
		{
			desc:  "all with addresses",
			value: "::",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
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
				}, nil
			},
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
		},
		{
			desc:  "single address match name",
			value: "1.1.1.1",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
					{Name: "1.1.1.1"},
					{Name: "2.2.2.2"},
				}, nil
			},
			interfaces: []pcap.Interface{
				{Name: "1.1.1.1"},
			},
		},
		{
			desc:  "single address match IP",
			value: "1.1.1.1",
			getInterfacesFn: func() ([]pcap.Interface, error) {
				return []pcap.Interface{
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
				}, nil
			},
			interfaces: []pcap.Interface{
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
			addr := NewAddr(c.value)
			addr.getInterfacesFn = c.getInterfacesFn

			ifs, err := addr.Interfaces()
			if err != nil {
				assert.ErrorContains(t, err, c.errContains)
			} else {
				assert.DeepEqual(t, ifs, c.interfaces)
			}
		})
	}
}
