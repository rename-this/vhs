package capture

import (
	"errors"
	"io"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/gramLabs/vhs/session"
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
			fn: func(_ *session.Context, i pcap.Interface) {},
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
			fn: func(ctx *session.Context, i pcap.Interface) {
				if i.Name == "111" {
					ctx.Errors <- errors.New("111")
				}
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			errs := make(chan error, 1)
			ctx, _, _ := session.NewContexts(nil, errs)

			c.listener.listenAll(ctx, c.fn)
			if c.errContains == "" {
				assert.Equal(t, 0, len(ctx.Errors))
			} else {
				assert.Equal(t, 1, len(ctx.Errors))
				assert.ErrorContains(t, <-ctx.Errors, c.errContains)
			}
		})
	}
}

type testPacketDataSource struct {
	idx  int
	data []string
}

func (tpds *testPacketDataSource) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	defer func() { tpds.idx++ }()
	if tpds.idx >= len(tpds.data) {
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
			listener: NewListener(&Capture{}),
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
