package httpx

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/go-errors/errors"
)

// Ensure Request implements the Message interface.
var _ Message = &Request{}

// Request represents an HTTP request.
type Request struct {
	ConnectionID     string         `json:"connection_id,omitempty"`
	ExchangeID       int64          `json:"exchange_id,omitempty"`
	Created          time.Time      `json:"created,omitempty"`
	Method           string         `json:"method,omitempty"`
	URL              *url.URL       `json:"url,omitempty"`
	Proto            string         `json:"proto,omitempty"`
	ProtoMajor       int            `json:"proto_major,omitempty"`
	ProtoMinor       int            `json:"proto_minor,omitempty"`
	Header           http.Header    `json:"header,omitempty"`
	MimeType         string         `json:"mimetype,omitempty"`
	PostForm         url.Values     `json:"postform,omitempty"`
	Cookies          []*http.Cookie `json:"cookies,omitempty"`
	Body             string         `json:"body,omitempty"`
	ContentLength    int64          `json:"content_length,omitempty"`
	TransferEncoding []string       `json:"transfer_encoding,omitempty"`
	Host             string         `json:"host,omitempty"`
	Trailer          http.Header    `json:"trailer,omitempty"`
	RemoteAddr       string         `json:"remote_addr,omitempty"`
	RequestURI       string         `json:"request_uri,omitempty"`
	Response         *Response      `json:"response,omitempty"`
	SessionID        string         `json:"session_id,omitempty"`
}

// GetConnectionID gets a connection ID.
func (r *Request) GetConnectionID() string { return r.ConnectionID }

// GetExchangeID gets an exchange ID.
func (r *Request) GetExchangeID() int64 { return r.ExchangeID }

// SetCreated sets the created timestamp
func (r *Request) SetCreated(created time.Time) { r.Created = created }

// SetSessionID sets the session ID
func (r *Request) SetSessionID(id string) { r.SessionID = id }

// NewRequest creates a new Request.
func NewRequest(b *bufio.Reader, connectionID string, exchangeID int64) (*Request, error) {
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil, errors.Errorf("failed to read request: %w", err)
	}

	defer req.Body.Close()

	var mimetype string
	ct := req.Header.Get("Content-type")
	if ct == "" {
		var reader io.Reader = req.Body
		b, _ := ioutil.ReadAll(reader)
		mimetype = http.DetectContentType(b)
	} else {
		mimetype, _, _ = mime.ParseMediaType(ct)
	}

	if mimetype == "application/x-www-form-urlencoded" {
		err = req.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read request body: %w", err)
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
		MimeType:         mimetype,
		PostForm:         req.PostForm,
		Cookies:          req.Cookies(),
		Body:             string(body),
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}, nil
}
