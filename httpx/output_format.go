package httpx

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rename-this/vhs/core"
)

// NewOutputFormat creates an HTTP output format.
func NewOutputFormat(ctx core.Context) (core.OutputFormat, error) {
	registerEnvelopes(ctx)
	return &outputFormat{
		in:       make(chan interface{}),
		complete: make(chan struct{}, 1),
	}, nil
}

type outputFormat struct {
	in       chan interface{}
	complete chan struct{}
}

func (o *outputFormat) In() chan<- interface{} {
	return o.in
}

func (o *outputFormat) Complete() <-chan struct{} {
	return o.complete
}

func (o *outputFormat) Init(ctx core.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "output_http").
		Logger()

	ctx.Logger.Debug().Msg("init")

	defer func() {
		o.complete <- struct{}{}
	}()

	var (
		once   sync.Once
		offset time.Duration
	)

	for {
		select {
		case n := <-o.in:
			switch r := n.(type) {
			case *Request:
				once.Do(func() {
					offset = time.Now().Sub(r.Created)
				})

				wait := r.Created.Add(offset).Sub(time.Now())
				go o.writeRequest(ctx, wait, w, r)
			case *Response:
				// Ignore for now.
			default:
				ctx.Errors <- errors.New("http output format: unknown type")
			}
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}

func (o *outputFormat) writeRequest(ctx core.Context, wait time.Duration, w io.Writer, r *Request) {
	time.Sleep(wait)

	if err := r.StdRequest().Write(w); err != nil {
		ctx.Errors <- fmt.Errorf("failed to write HTTP request: %w", err)
	}
}
