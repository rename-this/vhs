package tcp

import (
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/rename-this/vhs/capture"
	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

const (
	nilPayload       = "<nil>"
	wrongPayloadType = "<wrongtype>"
)

func newPacket(t *testing.T, data string, srcPort, dstPort uint16) gopacket.Packet {
	var (
		ipLayer = layers.IPv4{
			SrcIP:    net.ParseIP("0.0.0.0"),
			DstIP:    net.ParseIP("0.0.0.0"),
			Protocol: layers.IPProtocolTCP,
		}
		tcpLayer = layers.TCP{
			SrcPort: layers.TCPPort(srcPort),
			DstPort: layers.TCPPort(dstPort),
		}
		opts = gopacket.SerializeOptions{
			FixLengths: true,
		}
		buf = gopacket.NewSerializeBuffer()
	)

	err := gopacket.SerializeLayers(buf, opts, &ipLayer, &tcpLayer, gopacket.Payload(data))
	assert.NilError(t, err)

	p := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv4, gopacket.DecodeOptions{})

	f, ok := p.Layer(gopacket.LayerTypeDecodeFailure).(*gopacket.DecodeFailure)
	if ok {
		assert.NilError(t, f.Error())
	}

	return p
}

func newUpPacket(t *testing.T, data string) gopacket.Packet {
	return newPacket(t, data, 1111, 2222)
}

func newDownPacket(t *testing.T, data string) gopacket.Packet {
	return newPacket(t, data, 2222, 1111)
}

type testListener struct {
	packets chan gopacket.Packet
}

func newTestListener(t *testing.T, data []string) capture.Listener {
	l := &testListener{
		packets: make(chan gopacket.Packet, len(data)),
	}
	for _, d := range data {
		switch d {
		case nilPayload:
			l.packets <- nil
		case wrongPayloadType:
			l.packets <- gopacket.NewPacket([]byte{}, layers.LayerTypeARP, gopacket.DecodeOptions{})
		default:
			l.packets <- newPacket(t, d, 1111, 22222)
		}
	}
	return l
}

func (l *testListener) Packets() <-chan gopacket.Packet { return l.packets }
func (l *testListener) Listen(ctx session.Context)      {}
func (l *testListener) Close()                          {}

func TestRead(t *testing.T) {
	cfg := &session.Config{
		DebugPackets: true,
	}
	flowCfg := &session.FlowConfig{
		SourceDuration: 800 * time.Millisecond,
		TCPTimeout:     50 * time.Millisecond,
	}
	cases := []struct {
		desc     string
		cfg      *session.Config
		flowCfg  *session.FlowConfig
		listener capture.Listener
		data     []string
		out      []string
	}{
		{
			desc:    "nil",
			cfg:     cfg,
			flowCfg: flowCfg,
			data:    []string{nilPayload},
		},
		{
			desc:    "empty packet",
			cfg:     cfg,
			flowCfg: flowCfg,
			data:    []string{""},
		},

		{
			desc:    "wrong packet type",
			cfg:     cfg,
			flowCfg: flowCfg,
			data:    []string{wrongPayloadType},
		},
		{
			desc:    "one packet",
			cfg:     cfg,
			flowCfg: flowCfg,
			data:    []string{"aaa"},
			out: []string{
				"aaa",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs = make(chan error)
				ctx  = session.NewContexts(c.cfg, c.flowCfg, errs)
			)

			defer ctx.Cancel()

			source, err := NewSource(ctx)
			assert.NilError(t, err)

			s, ok := source.(*tcpSource)
			assert.Assert(t, ok)

			go s.read(ctx, func(string, bool) (*capture.Capture, error) {
				return nil, nil
			}, func(*capture.Capture) capture.Listener {
				return newTestListener(t, c.data)
			})

			// Allow time for flushing and pruning.
			time.Sleep(time.Second)

			if len(c.out) == 0 {
				return
			}

			r := <-s.Streams()
			defer r.Close()

			assert.Assert(t, r.Meta().SourceID != "")

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			out := string(b)
			for _, o := range c.out {
				assert.Assert(t, strings.Contains(out, o))
			}
		})
	}
}
