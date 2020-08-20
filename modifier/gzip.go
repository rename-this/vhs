package modifier

import (
	"compress/gzip"
	"io"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// NewGzipWriteCloser creates a new gzip write closer.
func NewGzipWriteCloser(_ *session.Context) (WriteCloser, error) {
	return &gzipWriteCloser{}, nil
}

type gzipWriteCloser struct{}

func (*gzipWriteCloser) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// NewGzipReadCloser creates a new gzip read closer.
func NewGzipReadCloser(_ *session.Context) (ReadCloser, error) {
	return &gzipReadCloser{}, nil
}

type gzipReadCloser struct{}

func (*gzipReadCloser) Wrap(r io.ReadCloser) (io.ReadCloser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Errorf("failed to create gzip reader: %w", err)
	}
	return gr, nil
}
