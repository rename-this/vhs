package output

import (
	"context"

	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/modifier"
	"github.com/gramLabs/vhs/output/sink"
)

// Pipe joins a format and sink.
type Pipe struct {
	Format    format.Format
	Sink      sink.Sink
	Modifiers modifier.Modifiers
	Errors    chan error
}

// NewPipe creates a pipe connecting a format and a sink.
func NewPipe(f format.Format, s sink.Sink, ms ...modifier.Modifier) *Pipe {
	return &Pipe{
		Format:    f,
		Sink:      s,
		Modifiers: ms,
	}
}

// Init starts the pipe.
func (p *Pipe) Init(ctx context.Context) {
	defer func() {
		if err := p.Sink.Close(); err != nil {
			p.Errors <- err
		}
	}()

	w, closeAll := p.Modifiers.Wrap(p.Sink)
	defer func() {
		if err := closeAll(); err != nil {
			p.Errors <- err
		}
	}()

	go func() {
		for err := range p.Format.Errors() {
			p.Errors <- err
		}
	}()

	p.Format.Init(ctx, w)
}

// Start starts the pipe.
func (p *Pipe) Write(n interface{}) {
	p.Format.In() <- n
}

// Pipes is a slice of pipes
type Pipes []*Pipe

func (pp Pipes) Write(n interface{}) {
	for _, p := range pp {
		// TODO(andrewhare): should we parallelize this?
		p.Write(n)
	}
}
