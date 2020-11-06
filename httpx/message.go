package httpx

import (
	"time"

	"github.com/gramLabs/vhs/envelope"
	"github.com/gramLabs/vhs/session"
)

// MessageType is the type of an HTTP message.
type MessageType byte

const (
	// TypeRequest is an HTTP request.
	TypeRequest = '0'
	// TypeResponse is an HTTP response.
	TypeResponse = '1'
)

// Message is an HTTP message.
type Message interface {
	GetConnectionID() string
	GetExchangeID() string
	SetCreated(time.Time)
	SetSessionID(string)
}

func registerEnvelopes(ctx session.Context) {
	ctx.Registry.Register(func() envelope.Kindify { return &Request{} })
	ctx.Registry.Register(func() envelope.Kindify { return &Response{} })
}
