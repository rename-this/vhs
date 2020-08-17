package testhelper

import (
	"fmt"
	"io"

	"github.com/gramLabs/vhs/session"
)

// Format is a test format.
type Format struct {
	in       chan interface{}
	buffered bool
}

// NewFormat creates a new format.
func NewFormat(buffered bool) *Format {
	return &Format{
		in:       make(chan interface{}),
		buffered: buffered,
	}
}

// Init initializes the format.
func (f *Format) Init(ctx *session.Context, w io.Writer) {
	if f.buffered {
		f.initBuffered(ctx, w)
	} else {
		f.initUnbuffered(ctx, w)
	}
}

func (f *Format) initBuffered(ctx *session.Context, w io.Writer) {
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

func (f *Format) initUnbuffered(ctx *session.Context, w io.Writer) {
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

// In gets the in channel.
func (f *Format) In() chan<- interface{} { return f.in }
