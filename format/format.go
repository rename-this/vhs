package format

import (
	"io"

	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// Input is an interface for formatting input
type Input interface {
	Init(*session.Context, *middleware.Middleware, io.ReadCloser) error
	Out() <-chan interface{}
}

// Output is an interface for formatting output
type Output interface {
	Init(*session.Context, io.Writer)
	In() chan<- interface{}
}
