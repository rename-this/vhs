package flow

import (
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

func newTestSource(_ session.Context) (Source, error) {
	return &testSource{
		streams: make(chan ioutilx.ReadCloserID),
	}, nil
}

type testSource struct {
	streams chan ioutilx.ReadCloserID
	data    []ioutilx.ReadCloserID
}

func (s *testSource) Streams() <-chan ioutilx.ReadCloserID { return s.streams }

func (s *testSource) Init(ctx session.Context) {
	for _, d := range s.data {
		s.streams <- d
	}
}
