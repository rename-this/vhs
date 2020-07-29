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
		out: make(chan io.Reader),
	}
}

type jsonFormat struct {
	in  chan interface{}
	out chan io.Reader
}

func (p *jsonFormat) In() chan<- interface{} { return p.in }
func (p *jsonFormat) Out() <-chan io.Reader  { return p.out }

func (p *jsonFormat) Init(_ context.Context) {
	for n := range p.in {
		p.out <- NewJSONReader(n)
	}
}

// NewJSONReader creates an io.Reader around a type.
func NewJSONReader(n interface{}) io.Reader {
	return &jsonReader{n: n}
}

type jsonReader struct {
	n interface{}
}

func (r *jsonReader) Read(p []byte) (int, error) {
	b, err := json.Marshal(r.n)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return copy(p, b), io.EOF
}
