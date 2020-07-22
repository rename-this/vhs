package session

import "github.com/google/uuid"

// Session is a collected set of recorded data
type Session struct {
	ID string
}

// New creates a new session.
func New() *Session {
	return &Session{
		ID: uuid.New().String(),
	}
}
