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
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
}

type jsonFormat struct {
	in  chan interface{}
	out chan interface{}
}

func (p *jsonFormat) In() chan<- interface{}  { return p.in }
func (p *jsonFormat) Out() <-chan interface{} { return p.out }

func (p *jsonFormat) Init(_ context.Context) {
	for n := range p.in {
		// TODO(andrewhare): Pool these json writers.
		p.out <- &jsonWriter{n: n}
	}
}

type jsonWriter struct {
	n interface{}
}

func (jw *jsonWriter) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(jw.n)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	n, err := w.Write(b)
	return int64(n), err
}
