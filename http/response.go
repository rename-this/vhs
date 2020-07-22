package http

import (
	"bufio"
	"fmt"
	"io/ioutil"
	_http "net/http"
)

// Ensure Response implements the Message interface.
var _ Message = &Response{}

// Response represents an HTTP response.
type Response struct {
	ConnectionID     string              `json:"connection_id,omitempty"`
	ExchangeID       int64               `json:"exchange_id"`
	Status           string              `json:"status,omitempty"`
	StatusCode       int                 `json:"status_code,omitempty"`
	Proto            string              `json:"proto,omitempty"`
	ProtoMajor       int                 `json:"proto_major,omitempty"`
	ProtoMinor       int                 `json:"proto_minor,omitempty"`
	Header           map[string][]string `json:"header,omitempty"`
	Body             string              `json:"body,omitempty"`
	ContentLength    int64               `json:"content_length,omitempty"`
	TransferEncoding []string            `json:"transfer_encoding,omitempty"`
	Close            bool                `json:"close,omitempty"`
	Uncompressed     bool                `json:"uncompressed,omitempty"`
	Trailer          map[string][]string `json:"trailer,omitempty"`
}

// GetConnectionID gets a connection ID.
func (r *Response) GetConnectionID() string { return r.ConnectionID }

// GetExchangeID gets an exchange ID.
func (r *Response) GetExchangeID() int64 { return r.ExchangeID }

// NewResponse creates a new Response.
func NewResponse(b *bufio.Reader, connectionID string, exchangeID int64) (*Response, error) {
	res, err := _http.ReadResponse(b, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		ConnectionID:     connectionID,
		ExchangeID:       exchangeID,
		Status:           res.Status,
		StatusCode:       res.StatusCode,
		Proto:            res.Proto,
		ProtoMajor:       res.ProtoMajor,
		ProtoMinor:       res.ProtoMinor,
		Header:           res.Header,
		Body:             string(body),
		ContentLength:    res.ContentLength,
		TransferEncoding: res.TransferEncoding,
		Close:            res.Close,
		Uncompressed:     res.Uncompressed,
		Trailer:          res.Trailer,
	}, nil
}
