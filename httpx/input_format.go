package httpx

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/tcp"
	"github.com/segmentio/ksuid"
)

const (
	exchangeIDBufferSize = 512
)

// NewInputFormat creates an HTTP input formatter.
func NewInputFormat(ctx core.Context) (core.InputFormat, error) {
	registerEnvelopes(ctx)
	return &inputFormat{
		out: make(chan interface{}),
	}, nil
}

type inputFormat struct {
	out chan interface{}
}

func (i *inputFormat) Init(ctx core.Context, m core.Middleware, streams <-chan core.InputReader) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "http_input_format").
		Logger()

	ctx.Logger.Debug().Msg("init")

	// As requests come in, an exchange ID will be generated
	// by the request parsing goroutine. The response goroutine
	// will read off this channel. Use the size to adjust backpressure
	// on reading requests, and we should consdier making this
	// a configuratble value.
	exchangeIDs := make(chan string, exchangeIDBufferSize)

	for rdr := range streams {
		go func(r core.InputReader) {
			defer func() {
				if err := r.Close(); err != nil {
					ctx.Errors <- fmt.Errorf("failed to close httpx input format: %w", err)
				}
			}()

			direction, ok := r.Meta().Get(tcp.MetaDirection)
			if !ok {
				ctx.Errors <- fmt.Errorf("failed to find direction for %s", r.Meta().SourceID)
				return
			}

			var (
				buf      = bufio.NewReader(r)
				sourceID = r.Meta().SourceID
			)

			switch direction {
			case tcp.DirectionUp:
				go func() {
					for {
						eID := ksuid.New().String()
						exchangeIDs <- eID

						req, err := NewRequest(buf, sourceID, eID, r.Meta())
						if isEOF(err) {
							return
						}
						if err != nil {
							ctx.Errors <- fmt.Errorf("failed to parse request: %w", err)
							continue
						}
						i.handle(ctx, m, TypeRequest, req)
					}
				}()
			case tcp.DirectionDown:
				go func() {
					for {
						eID := <-exchangeIDs

						res, err := NewResponse(buf, sourceID, eID, r.Meta())
						if isEOF(err) {
							return
						}
						if err != nil {
							ctx.Errors <- fmt.Errorf("failed to parse response: %w", err)
							continue
						}
						i.handle(ctx, m, TypeResponse, res)
					}
				}()
			default:
				ctx.Errors <- fmt.Errorf("invalid TCP direction: %s", direction)
				return
			}

			<-ctx.StdContext.Done()
		}(rdr)
	}
}

func (i *inputFormat) handle(ctx core.Context, m core.Middleware, t MessageType, msg Message) {
	msg.SetCreated(time.Now())
	msg.SetSessionID(ctx.SessionID)

	// By default, msgOut is the original message.
	// If middleware is defined, this will be
	// overwritten by the middleware output.
	msgOut := msg

	if m != nil {
		n, err := m.Exec(ctx, []byte{byte(t)}, msg)
		if err != nil {
			ctx.Errors <- fmt.Errorf("failed to run middleware: %w", err)
			return
		}
		msgOut = n.(Message)
		if ctx.Config.DebugHTTPMessages {
			ctx.Logger.Debug().Interface("msg", msgOut).Msg("message overwritten by middleware")
		} else {
			ctx.Logger.Debug().Msg("message overwritten by middleware")
		}
	}

	i.out <- msgOut
}

func (i *inputFormat) Out() <-chan interface{} { return i.out }

func isEOF(errs ...error) bool {
	for _, err := range errs {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return true
		}
	}
	return false
}
