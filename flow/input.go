package flow

import (
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// Input joins a source and a format with
// optional modifiers.
type Input struct {
	Source    Source
	Modifiers InputModifiers
	Format    InputFormat
}

// NewInput creates a new input.
func NewInput(s Source, mis InputModifiers, f InputFormat) *Input {
	return &Input{
		Source:    s,
		Modifiers: mis,
		Format:    f,
	}
}

// Init starts the input.
func (i *Input) Init(ctx *session.Context, m middleware.Middleware) {
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
				if err := closers.Close(); err != nil {
					ctx.Errors <- err
				}
			}()

			go i.Format.Init(ctx, m, r)
		case <-ctx.StdContext.Done():
			return
		}
	}
}
