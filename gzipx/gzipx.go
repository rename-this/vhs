package gzipx

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/rename-this/vhs/core"
)

// NewOutputModifier creates a new gzip output modifier.
func NewOutputModifier(_ core.Context) (core.OutputModifier, error) {
	return &outputModifier{}, nil
}

type outputModifier struct{}

func (*outputModifier) Wrap(w core.OutputWriter) (core.OutputWriter, error) {
	return &gzipWriter{
		Writer: gzip.NewWriter(w),
		parent: w,
	}, nil
}

type gzipWriter struct {
	*gzip.Writer
	parent io.Closer
}

func (w *gzipWriter) Close() error {
	return core.CloseSequentially(w.Writer, w.parent)
}

// NewInputModifier creates a new gzip input modifier.
func NewInputModifier(_ core.Context) (core.InputModifier, error) {
	return &inputModifier{}, nil
}

type inputModifier struct{}

func (*inputModifier) Wrap(r core.InputReader) (core.InputReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	return &gzipReader{
		Reader: gr,
		parent: r,
	}, nil
}

type gzipReader struct {
	*gzip.Reader
	parent core.InputReader
	meta   *core.Meta
}

func (r *gzipReader) Meta() *core.Meta {
	return r.parent.Meta()
}

func (r *gzipReader) Close() error {
	return core.CloseSequentially(r.Reader, r.parent)
}
