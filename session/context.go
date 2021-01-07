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
func NewContexts(cfg *Config, flowCfg *FlowConfig, errs chan error) Context {
	return NewContextsForWriter(cfg, flowCfg, errs, os.Stderr)
}

// NewContextsForWriter creates a new set of contexts
// with logs written to a specific writer.
func NewContextsForWriter(cfg *Config, flowCfg *FlowConfig, errs chan error, w io.Writer) Context {
	var (
		sessionID      = ksuid.New().String()
		registry       = envelope.NewRegistry()
		stdCtx, cancel = context.WithCancel(context.Background())
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
		FlowConfig: flowCfg,
		SessionID:  sessionID,
		StdContext: stdCtx,
		Cancel:     cancel,
		Errors:     errs,
		Logger:     logger,
		Registry:   registry,
	}
}

// Context is a context for a session.
type Context struct {
	Config     *Config
	FlowConfig *FlowConfig
	SessionID  string
	StdContext context.Context
	Cancel     context.CancelFunc
	Errors     chan error
	Logger     zerolog.Logger
	Registry   *envelope.Registry
}
