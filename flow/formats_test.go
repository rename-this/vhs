package flow

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
)

func newTestInputFormat(_ session.Context) (InputFormat, error) {
	return &testFormat{
		out: make(chan interface{}),
	}, nil
}

type testFormat struct {
	out chan interface{}
}

func (i *testFormat) Init(ctx session.Context, m middleware.Middleware, s <-chan InputReader) {
	for r := range s {
		go func() {
			defer func() {
				if err := r.Close(); err != nil {
					ctx.Errors <- err
				}
			}()

			go func() {
				s := bufio.NewScanner(r)
				for s.Scan() {
					ii, err := strconv.Atoi(strings.TrimSpace(s.Text()))
					if err != nil {
						ctx.Errors <- err
					}
					i.out <- ii
				}
			}()

			<-ctx.StdContext.Done()
		}()
	}
}

func (i *testFormat) Out() <-chan interface{} { return i.out }

func newTestOutputFormatNoErr(ctx session.Context) OutputFormat {
	o, _ := newTestOutputFormat(ctx)
	return o
}

func newTestOutputFormat(ctx session.Context) (OutputFormat, error) {
	return &testOutputFormat{
		in:       make(chan interface{}),
		buffered: ctx.FlowConfig.BufferOutput,
	}, nil
}

type testOutputFormat struct {
	in       chan interface{}
	buffered bool
}

func (f *testOutputFormat) Init(ctx session.Context, w io.Writer) {
	if f.buffered {
		f.initBuffered(ctx, w)
	} else {
		f.initUnbuffered(ctx, w)
	}
}

func (f *testOutputFormat) initBuffered(ctx session.Context, w io.Writer) {
	var total int
	for {
		select {
		case n := <-f.in:
			total += n.(int)
		case <-ctx.StdContext.Done():
			w.Write([]byte(fmt.Sprint(total)))
			return
		}
	}
}

func (f *testOutputFormat) initUnbuffered(ctx session.Context, w io.Writer) {
	for {
		select {
		case n := <-f.in:
			i := n.(int)
			w.Write([]byte(fmt.Sprint(i)))
		case <-ctx.StdContext.Done():
			return
		}
	}
}

func (f *testOutputFormat) In() chan<- interface{} { return f.in }
