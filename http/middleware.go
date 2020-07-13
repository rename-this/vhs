package http

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
func (m *Middleware) ExecMessage(t MessageType, r interface{}) (interface{}, error) {
	return m.Exec([]byte{byte(t)}, r)
}
