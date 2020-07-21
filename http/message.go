package http

import "time"

// Message is an HTTP message.
type Message interface {
	GetConnectionID() string
	GetExchangeID() int64
	SetCreated(time.Time)
}
