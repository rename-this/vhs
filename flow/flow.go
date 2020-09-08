package flow

import (
	"fmt"
	"time"

	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// Flow connects a input and one or more outputs.
type Flow struct {
	Input   *Input
	Outputs Outputs
}

// Run runs the flow.
func (f *Flow) Run(ctx, inputCtx, outputCtx *session.Context, m middleware.Middleware) {
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

	complete := time.After(inputCtx.Config.FlowDuration)
	for {
		select {
		case n := <-f.Input.Format.Out():
			f.Outputs.Write(n)
		case <-ctx.StdContext.Done():
			return
		case <-complete:
			return
		}
	}
}
