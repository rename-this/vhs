package session

import (
	"context"

	"github.com/google/uuid"
)

// NewContexts creates a new set of contexts.
func NewContexts(cfg *Config, errs chan error) (*Context, *Context, *Context) {
	var (
		sessionID = uuid.New().String()

		stdCtx1, cancel1 = context.WithCancel(context.Background())
		stdCtx2, cancel2 = context.WithCancel(context.Background())
		stdCtx3, cancel3 = context.WithCancel(context.Background())
	)

	return &Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx1,
			Cancel:     cancel1,
			Errors:     errs,
		},
		&Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx2,
			Cancel:     cancel2,
			Errors:     errs,
		},
		&Context{
			Config:     cfg,
			SessionID:  sessionID,
			StdContext: stdCtx3,
			Cancel:     cancel3,
			Errors:     errs,
		}
}

// Context is a context for an entire flow.
type Context struct {
	Config     *Config
	SessionID  string
	StdContext context.Context
	Cancel     context.CancelFunc
	Errors     chan error
}
