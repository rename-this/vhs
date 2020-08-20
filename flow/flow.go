package flow

import (
	"fmt"
	"time"

	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/pipe"
	"github.com/gramLabs/vhs/session"
)

// Flow connects input and output pipes.
type Flow struct {
	Input   *pipe.Input
	Outputs pipe.Outputs
}

// Run runs the flow.
func (f *Flow) Run(ctx, inputCtx, outputCtx *session.Context, m *middleware.Middleware) {
	go f.Input.Init(inputCtx, m)
	go f.Outputs.Init(outputCtx)

	defer func() {
		inputCtx.Cancel()
		fmt.Println("draining inputs...")
		time.Sleep(inputCtx.Config.InputDrainDuration)

		outputCtx.Cancel()
		fmt.Println("shutting down...")
		time.Sleep(inputCtx.Config.ShutdownDuration)

		f.Outputs.Close(outputCtx)
	}()

	for {
		select {
		case n := <-f.Input.Format.Out():
			f.Outputs.Write(n)
		case <-ctx.StdContext.Done():
			return
		case <-time.After(inputCtx.Config.FlowDuration):
			return
		}
	}
}
