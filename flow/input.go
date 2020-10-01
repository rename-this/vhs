package flow

import (
	"fmt"

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
func (i *Input) Init(ctx session.Context, m middleware.Middleware) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "input").
		Logger()

	ctx.Logger.Debug().Msg("init")

	go i.Source.Init(ctx)

	for {
		select {
		case rs := <-i.Source.Streams():
			r, closers, err := i.Modifiers.Wrap(rs)
			if err != nil {
				ctx.Errors <- errors.Errorf("failed to wrap source stream: %w", err)
				continue
			}

			ctx.Logger.Debug().Int("count", len(closers)).Msg("modifiers wrapped")

			go func() {
				i.Format.Init(ctx, m, r)
				ctx.Logger.Debug().Msg("closing modifiers")
				if err := closers.Close(); err != nil {
					ctx.Errors <- fmt.Errorf("failed to close input modifier: %w", err)
				}
			}()
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}
