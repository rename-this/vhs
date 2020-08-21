package pipe

import (
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/source"
)

// Input joins a source and a format with
// optional modifiers.
type Input struct {
	Source    source.Source
	Modifiers modifier.Inputs
	Format    format.Input
}

// NewInput creates a new input pipe.
func NewInput(s source.Source, mis modifier.Inputs, f format.Input) *Input {
	return &Input{
		Source:    s,
		Modifiers: mis,
		Format:    f,
	}
}

// Init starts the pipe.
func (i *Input) Init(ctx *session.Context, m *middleware.Middleware) {
	go i.Source.Init(ctx)

	for {
		select {
		case rs := <-i.Source.Streams():
			r, closers, err := i.Modifiers.Wrap(rs)
			if err != nil {
				ctx.Errors <- errors.Errorf("failed to wrap source stream: %w", err)
				continue
			}

			defer func() {
				ctx.Errors <- closers.Close()
			}()

			go i.Format.Init(ctx, m, r)
		case <-ctx.StdContext.Done():
			return
		}
	}
}
