package httpx

import (
	"context"
	"fmt"
	"io"

	"github.com/gramLabs/vhs/internal/mware"
)

// MessageType is the type of an HTTP message.
type MessageType byte

const (
	// TypeRequest is an HTTP request.
	TypeRequest = '0'
	// TypeResponse is an HTTP response.
	TypeResponse = '1'
)

// NewMiddleware creates a new HTTP Middleware.
func NewMiddleware(ctx context.Context, command string, stderr io.Writer) (*Middleware, error) {
	m, err := mware.New(ctx, command, stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mware: %w", err)
	}

	return &Middleware{M: m}, nil
}

// Middleware is HTTP middleware.
type Middleware struct {
	*mware.M
}

// ExecMessage executes a middleware request for an HTTP message.
func (m *Middleware) ExecMessage(t MessageType, msg Message) (Message, error) {
	n, err := m.Exec([]byte{byte(t)}, msg)
	if err != nil {
		return nil, err
	}
	if msgOut, ok := n.(Message); ok {
		return msgOut, nil
	}
	// TODO(andrewhare): ultraverbose logging here, middleware
	// returned something that wasn't a message. Not sure if
	// this is actually possible but best to avoid a panic until
	// we have tests around this.
	return nil, nil
}
