package observe

import (
	"errors"
	"io"
)

// ReadCloser is an observable io.ReadCloser.
type ReadCloser interface {
	io.ReadCloser
	EOF() <-chan struct{}
}

func NewReadCloser(r io.ReadCloser) ReadCloser {
	return &readCloser{
		rc: r,
		// Buffer this so that listening for EOF
		// is not a requirement.
		eof: make(chan struct{}, 1),
	}
}

type readCloser struct {
	rc  io.ReadCloser
	eof chan struct{}
}

func (r *readCloser) Read(p []byte) (int, error) {
	n, err := r.rc.Read(p)
	if errors.Is(err, io.EOF) {
		r.eof <- struct{}{}
	}
	return n, err
}

func (r *readCloser) EOF() <-chan struct{} {
	return r.eof
}

func (r *readCloser) Close() error {
	return r.rc.Close()
}
