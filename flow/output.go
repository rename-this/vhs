package flow

import (
	"fmt"
	"sync"

	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

// Output joins a format and sink with
// optional modifiers.
type Output struct {
	Format    OutputFormat
	Modifiers OutputModifiers
	Sink      Sink

	closersMu sync.RWMutex
	closers   ioutilx.Closers
}

// NewOutput creates an output connecting a format and a sink.
func NewOutput(f OutputFormat, mos OutputModifiers, s Sink) *Output {
	return &Output{
		Format:    f,
		Modifiers: mos,
		Sink:      s,
	}
}

// Init starts the output.
func (o *Output) Init(ctx session.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "output").
		Logger()

	ctx.Logger.Debug().Msg("init")

	w, closers, err := o.Modifiers.Wrap(o.Sink)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to wrap sink: %w", err)
		return
	}

	ctx.Logger.Debug().Int("count", len(closers)).Msg("modifiers wrapped")

	o.closersMu.Lock()
	o.closers = closers
	o.closersMu.Unlock()

	go o.Format.Init(ctx, w)
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
func (oo Outputs) Init(ctx session.Context) {
	for _, o := range oo {
		go o.Init(ctx)
	}
}

// Close closes all outputs.
func (oo Outputs) Close(ctx session.Context) {
	for _, o := range oo {
		o.closersMu.RLock()
		if err := o.closers.Close(); err != nil {
			ctx.Errors <- err
		}
		o.closersMu.RUnlock()
		if err := o.Sink.Close(); err != nil {
			ctx.Errors <- err
		}
	}
}
