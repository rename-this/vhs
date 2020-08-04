package output

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gramLabs/vhs/output/modifier"
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

type testFormat struct {
	in       chan interface{}
	errs     chan error
	buffered bool
}

func newTestFormat(buffered bool) *testFormat {
	return &testFormat{
		in:       make(chan interface{}),
		errs:     make(chan error),
		buffered: buffered,
	}
}

func (f *testFormat) Init(ctx context.Context, w io.Writer) {
	if f.buffered {
		f.initBuffered(ctx, w)
	} else {
		f.initUnbuffered(ctx, w)
	}
}

func (f *testFormat) initBuffered(ctx context.Context, w io.Writer) {
	var total int
	for {
		select {
		case n := <-f.in:
			total += n.(int)
		case <-ctx.Done():
			w.Write([]byte(fmt.Sprint(total)))
			return
		}
	}
}

func (f *testFormat) initUnbuffered(ctx context.Context, w io.Writer) {
	for {
		select {
		case n := <-f.in:
			i := n.(int)
			w.Write([]byte(fmt.Sprint(i)))
		case <-ctx.Done():
			return
		}
	}
}

func (f *testFormat) In() chan<- interface{} { return f.in }
func (f *testFormat) Errors() <-chan error   { return f.errs }

type doubleMod struct{}

func (*doubleMod) Wrap(w io.WriteCloser) io.WriteCloser {
	return modifier.NopWriteCloser(&double{w: w})
}

type double struct {
	w io.WriteCloser
}

func (r *double) Write(p []byte) (int, error) {
	return r.w.Write(append(p, p...))
}

func TestSink(t *testing.T) {
	cases := []struct {
		desc string
		p    *Pipe
		data []interface{}
		out  string
	}{
		{
			desc: "unbuffered",
			p:    NewPipe(newTestFormat(false), nil, &testSink{}),
			data: []interface{}{1, 2, 3},
			out:  `123`,
		},
		{
			desc: "buffered",
			p:    NewPipe(newTestFormat(true), nil, &testSink{}),
			data: []interface{}{1, 2, 3},
			out:  `6`,
		},
		{
			desc: "modifiers",
			p:    NewPipe(newTestFormat(false), []modifier.Modifier{&doubleMod{}, &doubleMod{}}, &testSink{}),
			data: []interface{}{1, 2, 3},
			out:  "111122223333",
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
