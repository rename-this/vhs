package flow

import (
	"fmt"

	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
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

	// This is an interim channel that allows
	// the streams to be wrapped before passed
	// to the format.
	streams := make(chan InputReader)

	go i.Format.Init(ctx, m, streams)
	go i.Source.Init(ctx)

	for {
		select {
		case rs := <-i.Source.Streams():
			r, err := i.Modifiers.Wrap(rs)
			if err != nil {
				ctx.Errors <- fmt.Errorf("failed to wrap source stream: %w", err)
				continue
			}
			streams <- r
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}
