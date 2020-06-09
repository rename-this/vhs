package capture

import (
	"errors"
	"io"
	"testing"

	"github.com/google/gopacket"
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

type testPacketDataSource struct {
	idx  int
	data []string
}

func (tpds *testPacketDataSource) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	defer func() { tpds.idx++ }()
	if len(tpds.data) == tpds.idx {
		return nil, gopacket.CaptureInfo{}, io.EOF
	}
	return []byte(tpds.data[tpds.idx]), gopacket.CaptureInfo{}, nil
}

func TestReadPackets(t *testing.T) {
	cases := []struct {
		desc     string
		listener *Listener
		source   *testPacketDataSource
		decoder  gopacket.Decoder
	}{
		{
			desc:     "reads to EOF",
			listener: NewListener(nil, 1),
			source: &testPacketDataSource{
				data: []string{
					"111",
					"1111",
					"11111",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			packets := c.listener.Packets()

			go c.listener.readPackets(c.source, c.decoder)
			for _, d := range c.source.data {
				p := <-packets
				assert.Equal(t, string(p.Data()), d)
			}
		})
	}
}
