package pipe

import (
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
)

// Output joins a format and sink with
// optional modifiers.
type Output struct {
	Format    format.Output
	Sink      sink.Sink
	Modifiers modifier.WriteClosers

	closers modifier.Closers
}

// NewOutput creates a pipe connecting a format and a sink.
func NewOutput(f format.Output, s sink.Sink, ms ...modifier.WriteCloser) *Output {
	return &Output{
		Format:    f,
		Sink:      s,
		Modifiers: ms,
	}
}

// Init starts the pipe.
func (o *Output) Init(ctx *session.Context) {
	w, closers, err := o.Modifiers.Wrap(o.Sink)
	if err != nil {
		ctx.Errors <- errors.Errorf("failed to wrap sink: %w", err)
		return
	}

	o.closers = closers

	go o.Format.Init(ctx, w)
}

// Start starts the pipe.
func (o *Output) Write(n interface{}) {
	o.Format.In() <- n
}

// Outputs is a slice of pipes
type Outputs []*Output

func (oo Outputs) Write(n interface{}) {
	for _, o := range oo {
		// TODO(andrewhare): should we parallelize this?
		o.Write(n)
	}
}

// Init initializes the outputs.
func (oo Outputs) Init(ctx *session.Context) {
	for _, o := range oo {
		go o.Init(ctx)
	}
}

// Close closes all outputs.
func (oo Outputs) Close(ctx *session.Context) {
	for _, o := range oo {
		ctx.Errors <- o.closers.Close()
		ctx.Errors <- o.Sink.Close()
	}
}
