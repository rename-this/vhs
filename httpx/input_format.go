package httpx

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// NewInputFormat creates an HTTP input formatter.
func NewInputFormat(_ session.Context) (flow.InputFormat, error) {
	return &inputFormat{
		out: make(chan interface{}),
	}, nil
}

type inputFormat struct {
	out chan interface{}
}

func (i *inputFormat) Init(ctx session.Context, m middleware.Middleware, r ioutilx.ReadCloserID) error {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "http_input_format").
		Logger()

	ctx.Logger.Debug().Msg("init")

	defer func() { ctx.Logger.Debug().Msg("closing") }()
	defer r.Close()

	var (
		id         = r.ID()
		exchangeID int64

		reqR, reqW = io.Pipe()
		resR, resW = io.Pipe()
		mw         = io.MultiWriter(reqW, resW)

		reqBuf = bufio.NewReader(reqR)
		resBuf = bufio.NewReader(resR)

		wg   sync.WaitGroup
		done = make(chan struct{})
	)

	go func() {
		for {
			var (
				req    *Request
				res    *Response
				reqErr error
				resErr error
			)

			wg.Add(2)

			go func() {
				req, reqErr = NewRequest(reqBuf, id, exchangeID)
				if req != nil {
					ctx.Logger.Debug().Msg("request received")
					i.handle(ctx, m, TypeRequest, req)
				}
				wg.Done()
			}()

			go func() {
				res, resErr = NewResponse(resBuf, id, exchangeID)
				if res != nil {
					ctx.Logger.Debug().Msg("response received")
					i.handle(ctx, m, TypeResponse, res)
				}
				wg.Done()
			}()

			wg.Wait()

			if isEOF(reqErr, resErr) {
				ctx.Logger.Debug().Msg("stream EOF")
				done <- struct{}{}
				return
			}

			exchangeID++
		}
	}()

	go func() {
		defer reqW.Close()
		defer resW.Close()
		io.Copy(mw, r)
	}()

	for {
		select {
		case <-ctx.StdContext.Done():
			return nil
		case <-done:
			return nil
		}
	}
}

func (i *inputFormat) handle(ctx session.Context, m middleware.Middleware, t MessageType, msg Message) {
	msg.SetCreated(time.Now())
	msg.SetSessionID(ctx.SessionID)

	// By default, msgOut is the original message.
	// If middleware is defined, this will be
	// overwritten by the middleware output.
	msgOut := msg

	if m != nil {
		n, err := m.Exec(ctx, []byte{byte(t)}, msg)
		if err != nil {
			ctx.Errors <- errors.Errorf("failed to run middleware: %w", err)
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
