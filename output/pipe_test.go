package output

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gramLabs/vhs/output/format"
	"gotest.tools/v3/assert"
)

type datum struct {
	B bool
}

type testSink struct {
	mu   sync.Mutex
	data []byte
}

func (s *testSink) Data() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (*testSink) Init(_ context.Context) {}

func (s *testSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, p...)

	return len(p), nil
}

func (*testSink) Close() error { return nil }

type formatUnbuffered struct {
	in  chan interface{}
	out chan io.Reader
}

func newFormatUnbuffered() *formatUnbuffered {
	return &formatUnbuffered{
		in:  make(chan interface{}),
		out: make(chan io.Reader),
	}
}

func (f *formatUnbuffered) Init(ctx context.Context) {
	for {
		select {
		case n := <-f.in:
			d := n.(*datum)
			d.B = true
			f.out <- format.NewJSONReader(d)
		case <-ctx.Done():
			return
		}
	}
}

func (f *formatUnbuffered) In() chan<- interface{} { return f.in }
func (f *formatUnbuffered) Out() <-chan io.Reader  { return f.out }

type formatBuffered struct {
	in  chan interface{}
	out chan io.Reader
}

func newFormatBuffered() *formatBuffered {
	return &formatBuffered{
		in:  make(chan interface{}),
		out: make(chan io.Reader),
	}
}

func (f *formatBuffered) Init(ctx context.Context) {
	var d []*datum
	for {
		select {
		case n := <-f.in:
			dd := n.(*datum)
			dd.B = true
			d = append(d, dd)
		case <-ctx.Done():
			f.out <- format.NewJSONReader(d)
			return
		}
	}
}

func (f *formatBuffered) In() chan<- interface{} { return f.in }
func (f *formatBuffered) Out() <-chan io.Reader  { return f.out }

func TestSink(t *testing.T) {
	cases := []struct {
		desc string
		p    *Pipe
		data []interface{}
		out  string
	}{
		{
			desc: "unbuffered",
			p:    NewPipe(newFormatUnbuffered(), &testSink{}),
			data: []interface{}{
				&datum{},
				&datum{},
				&datum{},
			},
			out: `{"B":true}{"B":true}{"B":true}`,
		},
		{
			desc: "buffered",
			p:    NewPipe(newFormatBuffered(), &testSink{}),
			data: []interface{}{
				&datum{},
				&datum{},
				&datum{},
			},
			out: `[{"B":true},{"B":true},{"B":true}]`,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			go c.p.Init(ctx)

			for _, d := range c.data {
				c.p.Write(d)
			}

			time.Sleep(500 * time.Millisecond)

			cancel()

			time.Sleep(100 * time.Millisecond)

			s := c.p.Sink.(*testSink)
			assert.DeepEqual(t, string(s.Data()), c.out)
		})
	}
}
