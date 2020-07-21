package http

import (
	"bufio"
	"fmt"
	"io/ioutil"
	_http "net/http"
	"net/url"
)

var _ Message = &Request{}

// Request represents an HTTP request.
type Request struct {
	ConnectionID     string              `json:"connection_id,omitempty"`
	ExchangeID       int64               `json:"exchange_id,omitempty"`
	Method           string              `json:"method,omitempty"`
	URL              *url.URL            `json:"url,omitempty"`
	Proto            string              `json:"proto,omitempty"`
	ProtoMajor       int                 `json:"proto_major,omitempty"`
	ProtoMinor       int                 `json:"proto_minor,omitempty"`
	Header           map[string][]string `json:"header,omitempty"`
	Body             string              `json:"body,omitempty"`
	ContentLength    int64               `json:"content_length,omitempty"`
	TransferEncoding []string            `json:"transfer_encoding,omitempty"`
	Host             string              `json:"host,omitempty"`
	Trailer          map[string][]string `json:"trailer,omitempty"`
	RemoteAddr       string              `json:"remote_addr,omitempty"`
	RequestURI       string              `json:"request_uri,omitempty"`
	Response         *Response           `json:"response,omitempty"`
}

// GetConnectionID gets a connection ID.
func (r *Request) GetConnectionID() string { return r.ConnectionID }

// GetExchangeID gets an exchange ID.
func (r *Request) GetExchangeID() int64 { return r.ExchangeID }

// NewRequest creates a new Request.
func NewRequest(b *bufio.Reader, connectionID string, exchangeID int64) (*Request, error) {
	req, err := _http.ReadRequest(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	return &Request{
		ConnectionID:     connectionID,
		ExchangeID:       exchangeID,
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           req.Header,
		Body:             string(body),
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}, nil
}
