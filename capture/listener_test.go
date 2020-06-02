package capture

import (
	"errors"
	"testing"

	"github.com/google/gopacket/pcap"
	"gotest.tools/v3/assert"
)

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
				Capture: &Capture{
					Interfaces: []pcap.Interface{
						{Name: "111"},
						{Name: "222"},
					},
				},
			},
			fn: func(i pcap.Interface) error {
				return nil
			},
		},
		{
			desc: "errors",
			listener: &Listener{
				Capture: &Capture{
					Interfaces: []pcap.Interface{
						{Name: "111"},
						{Name: "222"},
					},
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
