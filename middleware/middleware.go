package middleware

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"sync"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
	"github.com/rs/zerolog"
)

// Middleware is an interface that can modify objects.
type Middleware interface {
	Start() error
	Wait() error
	Close()
	Exec(session.Context, []byte, interface{}) (interface{}, error)
}

// New creates a new Middleware.
func New(ctx session.Context, command string) (Middleware, error) {
	if command == "" {
		return nil, nil
	}

	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "middleware").
		Str("command", command).
		Logger()

	cmd := exec.CommandContext(ctx.StdContext, command)
	cmd.Stderr = &logWrapper{
		l: ctx.Logger,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Errorf("failed to get stdin pipe: %w", err)
	}

	ctx.Logger.Debug().Msg("stdin pipe retrieved")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Errorf("failed to get stdout pipe: %w", err)
	}

	ctx.Logger.Debug().Msg("stdout pipe retrieved")

	return &mware{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

type mware struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.Writer
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// Start starts the middleware command and leaves it open for execution.
func (m *mware) Start() error {
	if err := m.cmd.Start(); err != nil {
		return errors.Errorf("failed to start middleware command: %w", err)
	}

	return nil
}

// Wait waits for the middleware to complete.
func (m *mware) Wait() error {
	return m.cmd.Wait()
}

// Close closes the middleware.
func (m *mware) Close() {
	m.stdout.Close()
}

// Exec executes the middleware for n. The header bytes are written
// directly before the payload (which is JSON serialized) separated
// by a single space.
func (m *mware) Exec(ctx session.Context, header []byte, n interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scanner == nil {
		m.scanner = bufio.NewScanner(m.stdout)
		ctx.Logger.Debug().Msg("scanner created")
	}

	if len(header) > 0 {
		m.stdin.Write(header)
		m.stdin.Write([]byte{' '})
	}

	err := json.NewEncoder(m.stdin).Encode(n)
	if err != nil {
		return nil, errors.Errorf("failed to encode: %w", err)
	}

	ctx.Logger.Debug().Msgf("successfully encoded to stdin with header %v", header)

	if m.scanner.Scan() {
		ctx.Logger.Debug().Msg("scanner value received")

		if err := json.Unmarshal(m.scanner.Bytes(), n); err != nil {
			return nil, errors.Errorf("failed to unmarshal: %w", err)
		}

		ctx.Logger.Debug().Msg("successfully unmarshaled scanner value")
		return n, nil
	}

	return nil, m.scanner.Err()
}

type logWrapper struct {
	l zerolog.Logger
}

func (w *logWrapper) Write(p []byte) (int, error) {
	w.l.Debug().Str("stderr", string(p)).Msg("middleware logging")
	return len(p), nil
}
