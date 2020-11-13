package flow

import (
	"strconv"
	"sync"

	"github.com/rename-this/vhs/session"
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

type testSinkInt struct {
	optCloseErr error
	mu          sync.Mutex
	data        []int
}

func (*testSinkInt) Init(_ session.Context) {}

func (s *testSinkInt) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		return -1, err
	}

	s.data = append(s.data, int(i))

	return len(p), nil
}

func (s *testSinkInt) Close() error { return s.optCloseErr }
