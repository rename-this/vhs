package format

import (
	"context"
	"io"
)

// Format is an interface for formatting output
type Format interface {
	Init(context.Context)
	In() chan<- interface{}
	Out() <-chan io.Reader
}
