package format

import (
	"context"
	"encoding/json"
	"io"
)

// NewJSON creates a JSON formatter.
func NewJSON() Format {
	return &jsonFormat{
		in: make(chan interface{}),
	}
}

type jsonFormat struct {
	in chan interface{}
}

func (p *jsonFormat) In() chan<- interface{} { return p.in }

func (p *jsonFormat) Init(_ context.Context, w io.Writer) {
	enc := json.NewEncoder(w)
	for n := range p.in {
		enc.Encode(n)
	}
}
