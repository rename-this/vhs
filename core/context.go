package core

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

// NewContext creates a new context.
func NewContext(cfg *Config, flowCfg *FlowConfig, errs chan error) Context {
	return NewContextForWriter(cfg, flowCfg, errs, os.Stderr)
}

// NewContextForWriter creates a new context
// with logs written to a specific writer.
func NewContextForWriter(cfg *Config, flowCfg *FlowConfig, errs chan error, w io.Writer) Context {
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
