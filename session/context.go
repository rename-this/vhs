package session

import (
	"context"
	"io"
	"os"

	"github.com/rename-this/vhs/envelope"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

const (
	// LoggerKeyComponent is the logger key for components.
	LoggerKeyComponent = "component"
)

// NewContexts creates a new set of contexts.
func NewContexts(cfg *Config, errs chan error) (Context, Context, Context) {
	return NewContextsForWriter(cfg, errs, os.Stderr)
}

// NewContextsForWriter creates a new set of contexts
// with logs written to a specific writer.
func NewContextsForWriter(cfg *Config, errs chan error, w io.Writer) (Context, Context, Context) {
	var (
		sessionID = ksuid.New().String()

		stdCtx1, cancel1 = context.WithCancel(context.Background())
		stdCtx2, cancel2 = context.WithCancel(context.Background())
		stdCtx3, cancel3 = context.WithCancel(context.Background())

		registry = envelope.NewRegistry()
	)

	var (
		logWriter = zerolog.ConsoleWriter{
			Out: w,
		}
		logger = zerolog.New(logWriter).With().
			Str("session_id", sessionID).
			Timestamp().
			Logger()
	)

	if cfg != nil && !cfg.Debug {
		logger = logger.Level(zerolog.ErrorLevel)
	}

	return Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx1,
			Cancel:     cancel1,
			Errors:     errs,
			Logger:     logger,
			Registry:   registry,
		},
		Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx2,
			Cancel:     cancel2,
			Errors:     errs,
			Logger:     logger,
			Registry:   registry,
		},
		Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx3,
			Cancel:     cancel3,
			Errors:     errs,
			Logger:     logger,
			Registry:   registry,
		}
}

// Context is a context for a session.
type Context struct {
	Config     *Config
	SessionID  string
	StdContext context.Context
	Cancel     context.CancelFunc
	Errors     chan error
	Logger     zerolog.Logger
	Registry   *envelope.Registry
}
