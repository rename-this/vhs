package gzipx

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/session"
)

// NewOutputModifier creates a new gzip output modifier.
func NewOutputModifier(_ session.Context) (flow.OutputModifier, error) {
	return &outputModifier{}, nil
}

type outputModifier struct{}

func (*outputModifier) Wrap(w flow.OutputWriter) (flow.OutputWriter, error) {
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
	return flow.CloseSequentially(w.Writer, w.parent)
}

// NewInputModifier creates a new gzip input modifier.
func NewInputModifier(_ session.Context) (flow.InputModifier, error) {
	return &inputModifier{}, nil
}

type inputModifier struct{}

func (*inputModifier) Wrap(r flow.InputReader) (flow.InputReader, error) {
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
	parent io.Closer
}

func (r *gzipReader) Close() error {
	return flow.CloseSequentially(r.Reader, r.parent)
}
