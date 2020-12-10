package httpx

import (
	"io"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewOutputFormat creates a new HTTP output format.
func NewOutputFormat(ctx session.Context) (flow.OutputFormat, error) {
	registerEnvelopes(ctx)
	return &outputFormat{
		in: make(chan interface{}),
	}, nil
}

type outputFormat struct {
	in chan interface{}
}

func (o *outputFormat) Init(ctx session.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "output_format").
		Logger()

	ctx.Logger.Debug().Msg("init")
}

func (o *outputFormat) In() chan<- interface{} {
	return o.in
}
