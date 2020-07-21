package http

// Exchange is an HTTP request and response pair.
type Exchange struct {
	ConnectionID string
	ExchangeID   int64
	Request      *Request
	Response     *Response
}
