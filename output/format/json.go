package format

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// NewJSON creates a JSON formatter.
func NewJSON(buffered bool) Format {
	return &jsonFormat{
		in:       make(chan interface{}),
		errs:     make(chan error),
		buffered: buffered,
	}
}

type jsonFormat struct {
	in       chan interface{}
	errs     chan error
	buffered bool
}

func (p *jsonFormat) In() chan<- interface{} { return p.in }
func (p *jsonFormat) Errors() <-chan error   { return p.errs }

func (p *jsonFormat) Init(ctx context.Context, w io.Writer) {
	enc := json.NewEncoder(w)

	if p.buffered {
		var buf []interface{}
		for {
			select {
			case n := <-p.in:
				buf = append(buf, n)
			case <-ctx.Done():
				if err := enc.Encode(buf); err != nil {
					p.errs <- fmt.Errorf("failed to encode buffer to JSON: %w", err)
				}
				return
			}
		}
	}

	for n := range p.in {
		if err := enc.Encode(n); err != nil {
			p.errs <- fmt.Errorf("failed to encode to JSON: %w", err)
		}
	}
}
