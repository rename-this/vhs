package source

import (
	"io"

	"github.com/gramLabs/vhs/session"
)

// Source is a data source that can be consumed
// by an input pipe.
type Source interface {
	Init(*session.Context)
	Session() *session.Session
	Streams() <-chan io.ReadCloser
}
