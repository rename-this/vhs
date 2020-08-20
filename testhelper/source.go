package testhelper

import (
	"io"

	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/source"
)

// NewSource creates a new  source.
func NewSource(_ *session.Context) (source.Source, error) {
	return &src{
		streams: make(chan io.ReadCloser),
	}, nil
}

type src struct {
	streams chan io.ReadCloser
}

func (s *src) Streams() <-chan io.ReadCloser { return s.streams }

func (s *src) Init(ctx *session.Context) {}
