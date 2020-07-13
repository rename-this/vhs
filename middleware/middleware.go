package middleware

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Middleware represents an open executable where data can
// be written to its stdin and output is read from the executable's
// stdout,
type Middleware struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.Writer
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// New creates a new HTTP Middleware.
func New(ctx context.Context, command string, stderr io.Writer) (*Middleware, error) {
	if command == "" {
		return &Middleware{}, nil
	}

	cmd := exec.CommandContext(ctx, command)
	cmd.Stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	return &Middleware{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

// Start starts the middleware command and leaves it open for execution.
func (m *Middleware) Start() error {
	if err := m.cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	return nil
}

// Close closes the middleware.
func (m *Middleware) Close() {
	m.stdout.Close()
}

// Exec executes the middleware for a given request.
func (m *Middleware) Exec(n interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scanner == nil {
		m.scanner = bufio.NewScanner(m.stdout)
	}

	err := json.NewEncoder(m.stdin).Encode(n)
	if err != nil {
		return nil, fmt.Errorf("failed to encode: %w", err)
	}

	if m.scanner.Scan() {
		if err := json.Unmarshal(m.scanner.Bytes(), n); err != nil {
			return nil, fmt.Errorf("failed to unmarshal: %w", err)
		}
		return n, nil
	}
	if err := m.scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan middleware stdout: %w", err)
	}

	return nil, nil
}
