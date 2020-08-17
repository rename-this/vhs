package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/gramLabs/vhs/config"
)

// NewContexts creates a new set of contexts.
func NewContexts(cfg *config.Config) (*Context, *Context, *Context) {
	var (
		errs      = make(chan error)
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
	Config     *config.Config
	SessionID  string
	StdContext context.Context
	Cancel     context.CancelFunc
	Errors     chan error
}
