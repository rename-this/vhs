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
	Modifiers modifier.Modifiers
	Sink      sink.Sink
	Errors    chan error
}

// NewPipe creates a pipe connecting a format and a sink.
func NewPipe(f format.Format, ms modifier.Modifiers, s sink.Sink) *Pipe {
	return &Pipe{
		Format:    f,
		Modifiers: ms,
		Sink:      s,
	}
}

// Init starts the pipe.
func (p *Pipe) Init(ctx context.Context) {
	w, closeAll := p.Modifiers.Wrap(p.Sink)
	defer closeAll()

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
