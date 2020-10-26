package flow

import (
	"io"

	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// InputFormat is an interface for formatting input
type InputFormat interface {
	Init(session.Context, middleware.Middleware, InputReader) error
	Out() <-chan interface{}
}

// OutputFormat is an interface for formatting output
type OutputFormat interface {
	Init(session.Context, io.Writer)
	In() chan<- interface{}
}
