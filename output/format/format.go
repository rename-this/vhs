package format

import "context"

// Format is an interface for formatting output
type Format interface {
	Init(context.Context)
	In() chan<- interface{}
	Out() <-chan interface{}
}
