package core

import "io"

// OutputFormat is an interface for formatting output
type OutputFormat interface {
	Init(Context, io.Writer)
	In() chan<- interface{}
	Complete() <-chan struct{}
}
