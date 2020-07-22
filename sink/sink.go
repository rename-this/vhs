package sink

import (
	"context"

	"github.com/gramLabs/vhs/session"
)

// Sink is a writable location for output.
type Sink interface {
	Init(context.Context, *session.Session)
	Write(interface{}) error
	Flush() error
}
