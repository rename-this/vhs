package flow

import (
	"sync"

	"github.com/gramLabs/vhs/session"
)

type testSink struct {
	optCloseErr error
	mu          sync.Mutex
	data        []byte
}

func (s *testSink) Data() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (*testSink) Init(_ session.Context) {}

func (s *testSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, p...)

	return len(p), nil
}

func (s *testSink) Close() error { return s.optCloseErr }
