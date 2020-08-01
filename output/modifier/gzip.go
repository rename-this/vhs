package modifier

import (
	"compress/gzip"
	"io"
)

var _ Modifier = &Gzip{}

// Gzip is a modifier that gzips.
type Gzip struct{}

// Wrap wraps a reader so it can gzip its contents.
func (*Gzip) Wrap(r io.WriteCloser) io.WriteCloser {
	return gzip.NewWriter(r)
}
