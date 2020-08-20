package testhelper

import (
	"fmt"
	"io"

	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/session"
)

// NewOutputFormatNoErr creates a new format without an erro.
func NewOutputFormatNoErr(ctx *session.Context) format.Output {
	o, _ := NewOutputFormat(ctx)
	return o
}

// NewOutputFormat creates a new format.
func NewOutputFormat(ctx *session.Context) (format.Output, error) {
	return &outputFormat{
		in:       make(chan interface{}),
		buffered: ctx.Config.BufferOutput,
	}, nil
}

type outputFormat struct {
	in       chan interface{}
	buffered bool
}

func (f *outputFormat) Init(ctx *session.Context, w io.Writer) {
	if f.buffered {
		f.initBuffered(ctx, w)
	} else {
		f.initUnbuffered(ctx, w)
	}
}

func (f *outputFormat) initBuffered(ctx *session.Context, w io.Writer) {
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

func (f *outputFormat) initUnbuffered(ctx *session.Context, w io.Writer) {
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

func (f *outputFormat) In() chan<- interface{} { return f.in }
