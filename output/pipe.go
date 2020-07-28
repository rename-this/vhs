package output

import (
	"context"
	"log"

	"github.com/gramLabs/vhs/output/format"
	"github.com/gramLabs/vhs/output/sink"
)

// Pipe joins a format and sink.
type Pipe struct {
	Format format.Format
	Sink   sink.Sink
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
	go p.Sink.Init(ctx)

	for n := range p.Format.Out() {
		if err := p.Sink.Write(n); err != nil {
			// TODO(andrewhare): figure out the best way to log this error
			log.Println(err)
		}
	}
}

// Start starts the pipe.
func (p *Pipe) Write(n interface{}) {
	p.Format.In() <- n
}
