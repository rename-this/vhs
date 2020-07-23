package sink

import (
	"context"
)

// Format is an interface for formatting output
type Format interface {
	Do(interface{}) (interface{}, error)
}

// Sink is a writable location for output.
type Sink interface {
	Init(context.Context, Format)
	Write(interface{}) error
	Flush() error
}
