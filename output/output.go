package output

import (
	"context"
	"log"
)

// Format is an interface for formatting output
type Format interface {
	Init(context.Context)
	In() chan<- interface{}
	Out() <-chan interface{}
}

// Sink is a writable location for output.
type Sink interface {
	Init(context.Context)
	Write(interface{}) error
}

// Pipe joins a format and sink.
type Pipe struct {
	Format Format
	Sink   Sink
}

// NewPipe creates a pipe connecting a format and a sink.
func NewPipe(format Format, sink Sink) *Pipe {
	return &Pipe{
		Format: format,
		Sink:   sink,
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
