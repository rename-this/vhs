package modifier

import (
	"compress/gzip"
	"io"

	"github.com/go-errors/errors"
)

var (
	_ WriteCloser = &GzipWriteCloser{}
	_ ReadCloser  = &GzipReadCloser{}
)

// GzipWriteCloser is an outout modifier that compresses.
type GzipWriteCloser struct{}

// Wrap wraps a writer so it can gzip its contents.
func (*GzipWriteCloser) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// GzipReadCloser is an input modifier that decompresses.
type GzipReadCloser struct{}

// Wrap wraps a reader so it can decompress its contents.
func (*GzipReadCloser) Wrap(r io.ReadCloser) (io.ReadCloser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Errorf("failed to create gzip reader: %w", err)
	}
	return gr, nil
}
