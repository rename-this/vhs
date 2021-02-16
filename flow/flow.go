package flow

import (
	"github.com/rename-this/vhs/core"
)

// Flow connects a input and one or more outputs.
type Flow struct {
	Input   *Input
	Outputs Outputs
}

// Run runs the flow.
func (f *Flow) Run(ctx core.Context, m core.Middleware) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "flow").
		Logger()

	ctx.Logger.Debug().Msg("running")

	go f.Input.Init(ctx, m)
	go f.Outputs.Init(ctx)

	defer func() {
		ctx.Cancel()

		f.Outputs.Drain(ctx)

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
