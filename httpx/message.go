package httpx

import (
	"time"
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
	GetExchangeID() int64
	SetCreated(time.Time)
	SetSessionID(string)
}
