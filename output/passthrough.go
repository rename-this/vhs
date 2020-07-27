package output

import "context"

// NewPassthrough is an empty format that passes all values unmodified
func NewPassthrough() Format {
	return &passthrough{
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
}

type passthrough struct {
	in  chan interface{}
	out chan interface{}
}

func (p *passthrough) Init(_ context.Context) {
	for n := range p.in {
		p.out <- n
	}
}

func (p *passthrough) In() chan<- interface{}  { return p.in }
func (p *passthrough) Out() <-chan interface{} { return p.out }
