package httpx

import (
	"bufio"
	"bytes"
	"io"
	"time"

	"github.com/go-errors/errors"
	"github.com/google/uuid"
	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/session"
)

// NewInputFormat creates an HTTP input formatter.
func NewInputFormat(_ *session.Context) (format.Input, error) {
	return &inputFormat{
		out: make(chan interface{}),
	}, nil
}

type inputFormat struct {
	out chan interface{}
}

func (i *inputFormat) Init(ctx *session.Context, m *middleware.Middleware, r io.ReadCloser) error {
	defer r.Close()

	var (
		resBuf bytes.Buffer
		tee    = io.TeeReader(r, &resBuf)

		reqReader = bufio.NewReader(tee)
		resReader = bufio.NewReader(&resBuf)

		exchangeID   int64
		connectionID = uuid.New().String()
	)

	for {
		i.handle(ctx, m, TypeRequest, func() (Message, error) {
			return NewRequest(reqReader, connectionID, exchangeID)
		})
		i.handle(ctx, m, TypeResponse, func() (Message, error) {
			return NewResponse(resReader, connectionID, exchangeID)
		})

		exchangeID++
	}
}

func (i *inputFormat) handle(ctx *session.Context, m *middleware.Middleware, t MessageType, parseMessage func() (Message, error)) {
	msg, err := parseMessage()
	if err != nil {
		// TODO(andrewhare): ultraverbose logging.
		return
	}

	msg.SetCreated(time.Now())
	msg.SetSessionID(ctx.SessionID)

	// By default, msgOut is the original message.
	// If middleware is defined, this will be
	// overwritten by the middleware output.
	var msgOut = msg

	if m != nil {
		n, err := m.Exec([]byte{byte(t)}, msg)
		if err != nil {
			ctx.Errors <- errors.Errorf("failed to run middleware: %w", err)
			return
		}
		msgOut = n.(Message)
	}

	i.out <- msgOut
}

func (i *inputFormat) Out() <-chan interface{} { return i.out }
