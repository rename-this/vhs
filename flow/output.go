package flow

import (
	"fmt"
	"sync"

	"github.com/rename-this/vhs/core"
)

// Output joins a format and sink with
// optional modifiers.
type Output struct {
	Format    core.OutputFormat
	Modifiers core.OutputModifiers
	Sink      core.Sink

	done chan struct{}
}

// NewOutput creates an output connecting a format and a sink.
func NewOutput(f core.OutputFormat, mods core.OutputModifiers, s core.Sink) *Output {
	return &Output{
		Format:    f,
		Modifiers: mods,
		Sink:      s,
		done:      make(chan struct{}),
	}
}

// Done returns a channel that signals when the output has completed.
func (o *Output) Done() <-chan struct{} {
	return o.done
}

// Init starts the output.
func (o *Output) Init(ctx core.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "output").
		Logger()

	ctx.Logger.Debug().Msg("init")

	w, err := o.Modifiers.Wrap(o.Sink)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to wrap sink: %w", err)
		return
	}

	defer func() {
		if err := w.Close(); err != nil {
			ctx.Errors <- fmt.Errorf("failed to close sink: %w", err)
		}
		o.done <- struct{}{}
	}()

	o.Format.Init(ctx, w)
}

// Write writes to the output.
func (o *Output) Write(n interface{}) {
	o.Format.In() <- n
}

// Outputs is a slice of output.
type Outputs []*Output

func (oo Outputs) Write(n interface{}) {
	for _, o := range oo {
		// TODO(andrewhare): should we parallelize this?
		o.Write(n)
	}
}

// Init initializes the outputs.
func (oo Outputs) Init(ctx core.Context) {
	for _, o := range oo {
		go o.Init(ctx)
	}
}

// Drain drains all outputs.
func (oo Outputs) Drain(ctx core.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "outputs").
		Logger()

	ctx.Logger.Debug().Msg("draining")

	var (
		wg  sync.WaitGroup
		num = len(oo)
	)

	wg.Add(num)

	for i, output := range oo {
		go func(ii int, o *Output) {
			<-o.Done()
			ctx.Logger.Debug().Msgf("output %d/%d drained", ii+1, num)
			wg.Done()
		}(i, output)
	}

	wg.Wait()
}
