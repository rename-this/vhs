package jsonx

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/session"
)

// NewOutputFormat creates a JSON output.
func NewOutputFormat(_ session.Context) (flow.OutputFormat, error) {
	return &outputFormat{
		in: make(chan interface{}),
	}, nil
}

type outputFormat struct {
	in chan interface{}
}

func (f *outputFormat) In() chan<- interface{} { return f.in }

func (f *outputFormat) Init(ctx session.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "output_json").
		Logger()

	ctx.Logger.Debug().Msg("init")

	enc := json.NewEncoder(w)

	ctx.Logger.Debug().Msg("encoder created")

	for n := range f.in {
		if err := enc.Encode(n); err != nil {
			ctx.Errors <- fmt.Errorf("failed to encode to JSON: %w", err)
		}
		ctx.Logger.Debug().Msg("value encoded")
	}
}

// NewBufferedOutputFormat creates a buffered JSON formatter.
func NewBufferedOutputFormat(_ session.Context) (flow.OutputFormat, error) {
	return &bufferedOutputFormat{
		in: make(chan interface{}),
	}, nil
}

type bufferedOutputFormat struct {
	in       chan interface{}
	buffered bool
}

func (f *bufferedOutputFormat) In() chan<- interface{} { return f.in }

func (f *bufferedOutputFormat) Init(ctx session.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "buffered_output_json").
		Logger()

	ctx.Logger.Debug().Msg("init")

	enc := json.NewEncoder(w)

	ctx.Logger.Debug().Msg("encoder created")

	var buf []interface{}
	for {
		select {
		case n := <-f.in:
			buf = append(buf, n)
			ctx.Logger.Debug().Msg("value buffered")
		case <-ctx.StdContext.Done():
			if err := enc.Encode(buf); err != nil {
				ctx.Errors <- fmt.Errorf("failed to encode buffer to JSON: %w", err)
			}
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}
