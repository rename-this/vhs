package gzipx

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

// NewOutputModifier creates a new gzip output modifier.
func NewOutputModifier(_ session.Context) (flow.OutputModifier, error) {
	return &outputModifier{}, nil
}

type outputModifier struct{}

func (*outputModifier) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// NewInputModifier creates a new gzip input modifier.
func NewInputModifier(_ session.Context) (flow.InputModifier, error) {
	return &inputModifier{}, nil
}

type inputModifier struct{}

func (*inputModifier) Wrap(r ioutilx.ReadCloserID) (ioutilx.ReadCloserID, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	return &gzipReaderID{
		Reader: gr,
		id:     r.ID(),
	}, nil
}

type gzipReaderID struct {
	*gzip.Reader
	id string
}

func (g *gzipReaderID) ID() string {
	return g.id
}
