package http

// Message is an HTTP message.
type Message interface {
	GetConnectionID() string
	GetExchangeID() int64
}
