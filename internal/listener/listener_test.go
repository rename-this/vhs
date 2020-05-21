package listener

import (
	"errors"
	"testing"

	"github.com/google/gopacket/pcap"
	"gotest.tools/v3/assert"
)

func TestNewListener(t *testing.T) {
	cases := []struct {
		desc        string
		addr        *Addr
		port        uint16
		errContains string
		interfaces  []pcap.Interface
	}{
		{
			desc: "addr interfaces error",
			addr: &Addr{
				getInterfacesFn: func() ([]pcap.Interface, error) {
					return nil, errors.New("111")
				},
			},
			errContains: "111",
		},
		{
			desc: "no errors",
			addr: &Addr{
				Type: AddrLoopback,
				getInterfacesFn: func() ([]pcap.Interface, error) {
					return []pcap.Interface{
						{Name: "111"},
						{Name: "222"},
						{Name: "333"},
					}, nil
				},
			},
			port: 1111,
			interfaces: []pcap.Interface{
				{Name: "111"},
				{Name: "222"},
				{Name: "333"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			l, err := NewListener(c.addr, c.port)
			if err != nil {
				assert.ErrorContains(t, err, c.errContains)
				return
			}

			assert.Equal(t, l.Port, c.port)
			assert.DeepEqual(t, l.Interfaces, c.interfaces)
		})
	}
}

func TestListenAll(t *testing.T) {
	cases := []struct {
		desc        string
		listener    *Listener
		fn          listenFn
		errContains string
	}{
		{
			desc: "no errors",
			listener: &Listener{
				Interfaces: []pcap.Interface{
					{Name: "111"},
					{Name: "222"},
				},
			},
			fn: func(i pcap.Interface) error {
				return nil
			},
		},
		{
			desc: "errors",
			listener: &Listener{
				Interfaces: []pcap.Interface{
					{Name: "111"},
					{Name: "222"},
				},
			},
			fn: func(i pcap.Interface) error {
				if i.Name == "111" {
					return errors.New("111")
				}
				return nil
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.listener.listenAll(c.fn)
			assert.ErrorContains(t, err, c.errContains)
		})
	}
}
