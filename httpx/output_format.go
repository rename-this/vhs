package httpx

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewOutputFormat creates an HTTP output format.
func NewOutputFormat(_ session.Context) (flow.OutputFormat, error) {
	return &outputFormat{
		in: make(chan interface{}),
	}, nil
}

type outputFormat struct {
	in chan interface{}
}

func (o *outputFormat) In() chan<- interface{} {
	return o.in
}

func (o *outputFormat) Init(ctx session.Context, w io.Writer) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "output_http").
		Logger()

	ctx.Logger.Debug().Msg("init")

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

func (o *outputFormat) writeRequest(ctx session.Context, wait time.Duration, w io.Writer, r *Request) {
	time.Sleep(wait)

	if err := r.StdRequest().Write(w); err != nil {
		ctx.Errors <- fmt.Errorf("failed to write HTTP request: %w", err)
	}
}
