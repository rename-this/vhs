package core

import "io"

// InputReader is an input reader.
type InputReader interface {
	io.ReadCloser
	Meta() *Meta
}
