package capture

import (
	"errors"
	"io"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

func TestNewHandle(t *testing.T) {
	cases := []struct {
		desc        string
		i           pcap.Interface
		activate    activateFn
		errContains string
	}{
		{
			desc: "no error",
			i:    pcap.Interface{Name: "111"},
			activate: func(inactive *pcap.InactiveHandle) (*pcap.Handle, error) {
				return pcap.OpenOffline("../testdata/200722_tcp_anon.pcapng")
			},
		},
		{
			desc: "error",
			i:    pcap.Interface{Name: "111"},
			activate: func(inactive *pcap.InactiveHandle) (*pcap.Handle, error) {
				return nil, errors.New("1111")
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{}, nil)
			l := NewListener(&Capture{})
			_, err := l.(*listener).newHandle(ctx, c.i, c.activate)
			if c.errContains == "" {
				assert.NilError(t, err)
				l.Close()
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}

func TestNewInactiveHandler(t *testing.T) {
	l := NewListener(&Capture{})
	_, err := l.(*listener).newInactiveHandler("111")

	assert.NilError(t, err)
}

type testPacketDataSource struct {
	idx  int
	data []string
	err  error
}

func (tpds *testPacketDataSource) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if tpds.err != nil {
		defer func() { tpds.err = nil }()
		return nil, gopacket.CaptureInfo{}, tpds.err
	}
	defer func() { tpds.idx++ }()
	if tpds.idx >= len(tpds.data) {
		return nil, gopacket.CaptureInfo{}, io.EOF
	}
	return []byte(tpds.data[tpds.idx]), gopacket.CaptureInfo{}, nil
}

func TestReadPackets(t *testing.T) {
	cases := []struct {
		desc     string
		listener Listener
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
		{
			desc:     "err",
			listener: NewListener(&Capture{}),
			source: &testPacketDataSource{
				data: []string{
					"---",
				},
				err: errors.New("111"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			packets := c.listener.Packets()

			ctx := session.NewContexts(&session.Config{DebugPackets: true}, &session.FlowConfig{}, nil)

			go c.listener.(*listener).readPackets(ctx, c.source, c.decoder)
			for _, d := range c.source.data {
				p := <-packets
				assert.Equal(t, string(p.Data()), d)
			}
			ctx.Cancel()
			c.listener.Close()
		})
	}
}
