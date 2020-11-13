package flow

import (
	"io"

	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
)

// InputFormat is an interface for formatting input
type InputFormat interface {
	Init(session.Context, middleware.Middleware, <-chan InputReader)
	Out() <-chan interface{}
}

// OutputFormat is an interface for formatting output
type OutputFormat interface {
	Init(session.Context, io.Writer)
	In() chan<- interface{}
}
