package flow

import (
	"github.com/rename-this/vhs/session"
)

func newTestSource(_ session.Context) (Source, error) {
	return &testSource{
		streams: make(chan InputReader),
	}, nil
}

type testSource struct {
	streams chan InputReader
	data    []InputReader
}

func (s *testSource) Streams() <-chan InputReader {
	return s.streams
}

func (s *testSource) Init(ctx session.Context) {
	for _, d := range s.data {
		s.streams <- d
	}
	close(s.streams)
}
