package httpx

import (
	"bufio"
	"bytes"
	"io"
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
		id = r.ID()

		resBuf bytes.Buffer
		tee    = io.TeeReader(r, &resBuf)

		reqReader = bufio.NewReader(tee)
		resReader = bufio.NewReader(&resBuf)

		exchangeID int64
	)

	for {
		select {
		case <-ctx.StdContext.Done():
			return nil
		default:
			req, err := NewRequest(reqReader, id, exchangeID)
			if req != nil {
				ctx.Logger.Debug().Msg("request received")
				i.handle(ctx, m, TypeRequest, req)
			}

			res, err := NewResponse(resReader, id, exchangeID)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				ctx.Logger.Debug().Msg("EOF")
				return nil
			} else if res != nil {
				ctx.Logger.Debug().Msg("response received")
				i.handle(ctx, m, TypeResponse, res)
			} else {
				ctx.Logger.Debug().AnErr("err", err).Msg("bad resp")
			}

			exchangeID++
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
