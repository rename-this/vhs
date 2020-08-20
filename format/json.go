package format

import (
	"encoding/json"
	"io"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// NewJSON creates a JSON output.
func NewJSON(_ *session.Context) (Output, error) {
	return &jsonFormat{
		in:       make(chan interface{}),
		buffered: false,
	}, nil
}

// NewJSONBuffered creates a buffered JSON formatter.
func NewJSONBuffered(_ *session.Context) (Output, error) {
	return &jsonFormat{
		in:       make(chan interface{}),
		buffered: true,
	}, nil
}

type jsonFormat struct {
	in       chan interface{}
	buffered bool
}

func (p *jsonFormat) In() chan<- interface{} { return p.in }

func (p *jsonFormat) Init(ctx *session.Context, w io.Writer) {
	enc := json.NewEncoder(w)

	if p.buffered {
		var buf []interface{}
		for {
			select {
			case n := <-p.in:
				buf = append(buf, n)
			case <-ctx.StdContext.Done():
				if err := enc.Encode(buf); err != nil {
					ctx.Errors <- errors.Errorf("failed to encode buffer to JSON: %w", err)
				}
				return
			}
		}
	}

	for n := range p.in {
		if err := enc.Encode(n); err != nil {
			ctx.Errors <- errors.Errorf("failed to encode to JSON: %w", err)
		}
	}
}
