package output

import (
	"context"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type data struct {
	D []*datum
}

type datum struct {
	B bool
}

type testSink struct {
	mu        sync.Mutex
	data      *data
	dataSlice []*datum
}

func (s *testSink) Data() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dataSlice != nil {
		return s.dataSlice
	}
	return s.data
}

func (*testSink) Init(_ context.Context) {}

func (s *testSink) Write(n interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dataSlice != nil {
		s.dataSlice = append(s.dataSlice, n.(*datum))
	} else {
		// panic(fmt.Sprintf("%#v", n))
		s.data = n.(*data)
	}

	return nil
}

type formatUnbuffered struct {
	in  chan interface{}
	out chan interface{}
}

func newFormatUnbuffered() *formatUnbuffered {
	return &formatUnbuffered{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
}

func (f *formatUnbuffered) Init(ctx context.Context) {
	for {
		select {
		case n := <-f.in:
			d := n.(*datum)
			d.B = true
			f.out <- d
		case <-ctx.Done():
			return
		}
	}
}

func (f *formatUnbuffered) In() chan<- interface{}  { return f.in }
func (f *formatUnbuffered) Out() <-chan interface{} { return f.out }

type formatBuffered struct {
	in  chan interface{}
	out chan interface{}
}

func newFormatBuffered() *formatBuffered {
	return &formatBuffered{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
}

func (f *formatBuffered) Init(ctx context.Context) {
	d := &data{}
	for {
		select {
		case n := <-f.in:
			dd := n.(*datum)
			dd.B = true
			d.D = append(d.D, dd)
		case <-ctx.Done():
			f.out <- d
		}
	}
}

func (f *formatBuffered) In() chan<- interface{}  { return f.in }
func (f *formatBuffered) Out() <-chan interface{} { return f.out }

func TestSink(t *testing.T) {
	cases := []struct {
		desc string
		p    *Pipe
		data []interface{}
		out  interface{}
	}{
		{
			desc: "unbuffered",
			p:    NewPipe(newFormatUnbuffered(), &testSink{dataSlice: []*datum{}}),
			data: []interface{}{
				&datum{},
				&datum{},
				&datum{},
			},
			out: []*datum{
				{B: true},
				{B: true},
				{B: true},
			},
		},
		{
			desc: "buffered",
			p:    NewPipe(newFormatBuffered(), &testSink{}),
			data: []interface{}{
				&datum{},
				&datum{},
				&datum{},
			},
			out: &data{
				D: []*datum{
					{B: true},
					{B: true},
					{B: true},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			go c.p.Init(ctx)

			for _, d := range c.data {
				c.p.Write(d)
			}

			cancel()

			time.Sleep(100 * time.Millisecond)

			s := c.p.Sink.(*testSink)
			assert.DeepEqual(t, s.Data(), c.out)
		})
	}
}
