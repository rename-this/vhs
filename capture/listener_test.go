package capture

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/gramLabs/vhs/session"
	"gotest.tools/v3/assert"
)

// This code is not supposed to be threadsafe.
// This allows this test to run with the race detector.
var listenMu sync.Mutex

func TestListen(t *testing.T) {
	listenMu.Lock()
	defer listenMu.Unlock()

	ctx, _, _ := session.NewContexts(nil, nil)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, time.Now().UnixNano())
	}))

	go func() {
		defer s.Close()
		var (
			client = s.Client()
			ticker = time.Tick(time.Second)
		)
		for {
			select {
			// Send requests once per second until
			// the context is canceled.
			case <-ticker:
				_, err := client.Get(s.URL)
				assert.NilError(t, err)
			case <-ctx.StdContext.Done():
				return
			}
		}
	}()

	c, err := NewCapture(strings.TrimLeft(s.URL, "http://"), true)
	assert.NilError(t, err)

	l := NewListener(c)
	l.Listen(ctx)

	for i := 0; i < 3; i++ {
		<-l.Packets()
	}

	ctx.Cancel()
}

func TestNewHandle(t *testing.T) {
	cases := []struct {
		desc string
		i    pcap.Interface
		fail bool
	}{
		{
			desc: "empty interface",
			fail: true,
		},
		{
			desc: "bad interface",
			i: pcap.Interface{
				Name: "111",
			},
			fail: true,
		},
		{
			desc: "good interface",
			i: pcap.Interface{
				Name: "eth0",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				l    = &listener{Capture: &Capture{}}
				_, _ = l.newHandle(c.i)
			)
			defer l.Close()
			if c.fail {
				assert.Equal(t, len(l.handles), 0)
			} else {
				assert.Equal(t, len(l.handles), 1)
			}
		})
	}
}

type testPacketDataSource struct {
	idx  int
	data []string
	err  error
}

func (tpds *testPacketDataSource) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	if tpds.err != nil {
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
				err: errors.New("111"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			packets := c.listener.Packets()

			ctx, _, _ := session.NewContexts(nil, nil)

			go c.listener.(*listener).readPackets(ctx, c.source, c.decoder)
			for _, d := range c.source.data {
				p := <-packets
				assert.Equal(t, string(p.Data()), d)
			}
		})
	}
}
