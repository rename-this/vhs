package modifier

import (
	"compress/gzip"
	"io"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// NewGzipOutput creates a new gzip write closer.
func NewGzipOutput(_ *session.Context) (Output, error) {
	return &gzipOutput{}, nil
}

type gzipOutput struct{}

func (*gzipOutput) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// NewGzipInput creates a new gzip read closer.
func NewGzipInput(_ *session.Context) (Input, error) {
	return &gzipInput{}, nil
}

type gzipInput struct{}

func (*gzipInput) Wrap(r io.ReadCloser) (io.ReadCloser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Errorf("failed to create gzip reader: %w", err)
	}
	return gr, nil
}
