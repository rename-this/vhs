package middleware

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"sync"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// Middleware represents an open executable where data can
// be written to its stdin and output is read from the executable's
// stdout, It is designed to be wrapped by protocol-specific middleware.
type Middleware struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.Writer
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// New creates a new Mware.
func New(ctx *session.Context, command string, stderr io.Writer) (*Middleware, error) {
	if command == "" {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx.StdContext, command)
	cmd.Stderr = stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Errorf("failed to get stdout pipe: %w", err)
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
		return errors.Errorf("failed to run command: %w", err)
	}

	return nil
}

// Close closes the middleware.
func (m *Middleware) Close() {
	m.stdout.Close()
}

// Exec executes the middleware for n. The header bytes are written
// directly before the payload (which is JSON serialized) separated
// by a single space.
func (m *Middleware) Exec(header []byte, n interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scanner == nil {
		m.scanner = bufio.NewScanner(m.stdout)
	}

	if len(header) > 0 {
		m.stdin.Write(header)
		m.stdin.Write([]byte{' '})
	}

	err := json.NewEncoder(m.stdin).Encode(n)
	if err != nil {
		return nil, errors.Errorf("failed to encode: %w", err)
	}

	if m.scanner.Scan() {
		if err := json.Unmarshal(m.scanner.Bytes(), n); err != nil {
			return nil, errors.Errorf("failed to unmarshal: %w", err)
		}
		return n, nil
	}
	if err := m.scanner.Err(); err != nil {
		return nil, errors.Errorf("failed to scan middleware stdout: %w", err)
	}

	return nil, nil
}
