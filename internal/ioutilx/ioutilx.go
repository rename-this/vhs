package ioutilx

import (
	"io"

	"go.uber.org/multierr"
)

// Closers is a slice of io.Closer
type Closers []io.Closer

// Close closes all closers.
func (cs Closers) Close() error {
	var errs error
	for _, c := range cs {
		if err := c.Close(); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopWriteCloser{Writer: w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

// ReadCloserID is an identifiable ReadCloser.
type ReadCloserID interface {
	io.ReadCloser
	ID() string
}

// NopReadCloserID returns a ReadCloserID with a empty ID method wrapping
// the provided io.ReadCloser r.
func NopReadCloserID(r io.ReadCloser) ReadCloserID {
	return nopReadCloserID{ReadCloser: r}
}

type nopReadCloserID struct {
	io.ReadCloser
}

func (nopReadCloserID) ID() string { return "" }
