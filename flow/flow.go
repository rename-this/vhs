package flow

import (
	"time"

	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
)

// Flow connects a input and one or more outputs.
type Flow struct {
	Input   *Input
	Outputs Outputs
}

// Run runs the flow.
func (f *Flow) Run(ctx, inputCtx, outputCtx session.Context, m middleware.Middleware) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "flow").
		Logger()

	ctx.Logger.Debug().Msg("running")

	go f.Input.Init(inputCtx, m)
	go f.Outputs.Init(outputCtx)

	defer func() {
		inputCtx.Cancel()
		ctx.Logger.Debug().Dur("dur", inputCtx.Config.InputDrainDuration).Msg("draining inputs")
		time.Sleep(inputCtx.Config.InputDrainDuration)

		outputCtx.Cancel()
		ctx.Logger.Debug().Dur("dur", inputCtx.Config.ShutdownDuration).Msg("shutting down")
		time.Sleep(inputCtx.Config.ShutdownDuration)

		ctx.Logger.Debug().Msg("shutdown complete")
	}()

	complete := time.After(inputCtx.Config.FlowDuration)
	for {
		select {
		case n := <-f.Input.Format.Out():
			go f.Outputs.Write(n)
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		case <-complete:
			ctx.Logger.Debug().Msg("complete")
			return
		}
	}
}
