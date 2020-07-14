package http

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	_http "net/http"
)

// Response represents an HTTP response.
type Response struct {
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

// NewResponse creates a new Response.
func NewResponse(b *bufio.Reader) (*Response, error) {
	res, err := _http.ReadResponse(b, nil)
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
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
