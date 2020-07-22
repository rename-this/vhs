package http

import (
	"time"

	"github.com/gramLabs/vhs/session"
)

// Message is an HTTP message.
type Message interface {
	GetConnectionID() string
	GetExchangeID() int64
	SetCreated(time.Time)
	SetSession(*session.Session)
}
