package http

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Middleware represents an open executable where data can
// be written to its stdin and output is read from the executable's
// stdout,
type Middleware struct {
	Stderr io.ReadCloser

	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.Writer
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// NewMiddleware creates a new HTTP Middleware.
func NewMiddleware(ctx context.Context, command string, stderr io.Writer) (*Middleware, error) {
	if command == "" {
		return &Middleware{}, nil
	}

	var (
		parts = strings.Split(command, " ")
		cmd   = exec.CommandContext(ctx, parts[0], parts[1:]...)
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	cmd.Stderr = stderr

	return &Middleware{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

// Start starts the middleware command and leaves it open for execution.
func (m *Middleware) Start() error {
	if m.cmd == nil {
		return nil
	}

	err := m.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start middleware: %w", err)
	}

	if err := m.cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait on command: %w", err)
	}

	return nil
}

// Close closes the middleware.
func (m *Middleware) Close() {
	m.Stderr.Close()
	m.stdout.Close()
}

// Exec executes the middleware for a given request.
func (m *Middleware) Exec(req *Request) (*Request, error) {
	if m.stdin == nil {
		return req, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scanner == nil {
		m.scanner = bufio.NewScanner(m.stdout)
	}

	err := json.NewEncoder(m.stdin).Encode(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	if m.scanner.Scan() {
		var outReq Request
		if err := json.Unmarshal(m.scanner.Bytes(), &outReq); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request: %w", err)
		}
		return &outReq, nil
	}
	if err := m.scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan middleware stdout: %w", err)
	}

	return nil, nil
}
