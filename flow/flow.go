package flow

import (
	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
)

// Flow connects a input and one or more outputs.
type Flow struct {
	Input   *Input
	Outputs Outputs
}

// Run runs the flow.
func (f *Flow) Run(ctx session.Context, m middleware.Middleware) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "flow").
		Logger()

	ctx.Logger.Debug().Msg("running")

	go f.Input.Init(ctx, m)
	go f.Outputs.Init(ctx)

	defer func() {
		ctx.Cancel()

		f.Outputs.Close(ctx)

		ctx.Logger.Debug().Msg("complete")
	}()

	for {
		select {
		case n := <-f.Input.Format.Out():
			f.Outputs.Write(n)
		case <-f.Input.Done():
			return
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}
