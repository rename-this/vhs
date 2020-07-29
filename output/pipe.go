package output

import (
	"context"
	"io"

	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/sink"
)

// Pipe joins a format and sink.
type Pipe struct {
	Format format.Format
	Sink   sink.Sink
	Errors chan error
}

// NewPipe creates a pipe connecting a format and a sink.
func NewPipe(f format.Format, s sink.Sink) *Pipe {
	return &Pipe{
		Format: f,
		Sink:   s,
	}
}

// Init starts the pipe.
func (p *Pipe) Init(ctx context.Context) {
	go p.Format.Init(ctx)

	for r := range p.Format.Out() {
		if _, err := io.Copy(p.Sink, r); err != nil {
			p.Errors <- err
		}
	}
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
