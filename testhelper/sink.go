package testhelper

import (
	"sync"

	"github.com/gramLabs/vhs/session"
)

// Sink is a test sink.
type Sink struct {
	mu   sync.Mutex
	data []byte
}

// Data gets the data.
func (s *Sink) Data() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// Init is a no-op.
func (*Sink) Init(_ *session.Context) {}

// Write writes to the sink.
func (s *Sink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, p...)

	return len(p), nil
}

// Close is a no-op/
func (*Sink) Close() error { return nil }
