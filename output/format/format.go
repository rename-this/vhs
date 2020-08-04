package format

import (
	"context"
	"io"
)

// Format is an interface for formatting output
type Format interface {
	Init(context.Context, io.Writer)
	In() chan<- interface{}
	Errors() <-chan error
}
