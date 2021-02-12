package core

import "io"

// OutputWriter is an output writer.
type OutputWriter interface {
	io.WriteCloser
}
