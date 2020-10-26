package flow

import (
	"github.com/gramLabs/vhs/session"
)

// Source is a data source that can be consumed
// by an input pipe.
type Source interface {
	Init(session.Context)
	Streams() <-chan InputReader
}
