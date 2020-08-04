package format

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// NewJSON creates a JSON formatter.
func NewJSON() Format {
	return &jsonFormat{
		in:   make(chan interface{}),
		errs: make(chan error),
	}
}

type jsonFormat struct {
	in   chan interface{}
	errs chan error
}

func (p *jsonFormat) In() chan<- interface{} { return p.in }
func (p *jsonFormat) Errors() <-chan error   { return p.errs }

func (p *jsonFormat) Init(_ context.Context, w io.Writer) {
	enc := json.NewEncoder(w)
	for n := range p.in {
		if err := enc.Encode(n); err != nil {
			p.errs <- fmt.Errorf("failed to encode to JSON: %w", err)
		}
	}
}
