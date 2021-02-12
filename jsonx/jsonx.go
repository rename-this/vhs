package jsonx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/envelope"
)

// NewInputFormat creates a new JSON input format.
func NewInputFormat(_ core.Context) (core.InputFormat, error) {
	return &inputFormat{
		out: make(chan interface{}),
	}, nil
}

type inputFormat struct {
	out chan interface{}
}

func (i *inputFormat) Out() <-chan interface{} {
	return i.out
}

func (i *inputFormat) Init(ctx core.Context, m core.Middleware, streams <-chan core.InputReader) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "json_input_format").
		Logger()

	ctx.Logger.Debug().Msg("init")

	for rdr := range streams {
		go func(r core.InputReader) {
			defer func() {
				if err := r.Close(); err != nil {
					ctx.Errors <- fmt.Errorf("failed to close JSON input format: %w", err)
				}
			}()

			dec := json.NewDecoder(r)
			for {
				select {
				case <-ctx.StdContext.Done():
					ctx.Logger.Debug().Msg("context canceled")
					return
				default:
					n, err := ctx.Registry.DecodeJSON(dec)
					if errors.Is(err, io.EOF) {
						return
					}
					if err != nil {
						ctx.Errors <- fmt.Errorf("failed to decode input JSON: %w", err)
						continue
					}

					ctx.Logger.Debug().
						Interface("n", n).
						Str("n_type", reflect.TypeOf(n).String()).
						Msg("received JSON input")

					i.out <- n
				}
			}
		}(rdr)
	}
}

// NewOutputFormat creates a JSON output.
func NewOutputFormat(_ core.Context) (core.OutputFormat, error) {
	return &outputFormat{
		in: make(chan interface{}),
	}, nil
}

type outputFormat struct {
	in chan interface{}
}

func (f *outputFormat) In() chan<- interface{} { return f.in }

func (f *outputFormat) Init(ctx core.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "output_json").
		Logger()

	ctx.Logger.Debug().Msg("init")

	enc := json.NewEncoder(w)

	ctx.Logger.Debug().Msg("encoder created")

	for {
		select {
		case n := <-f.in:
			if namer, ok := n.(envelope.Kindify); ok {
				n = envelope.New(namer)
			}
			if err := enc.Encode(n); err != nil {
				ctx.Errors <- fmt.Errorf("failed to encode to JSON: %w", err)
			}
			ctx.Logger.Debug().Msg("value encoded")
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}
