package httpx

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/envelope"
	"github.com/rename-this/vhs/tcp"
)

// Ensure Request implements the Message interface.
var _ Message = &Request{}

// Request represents an HTTP request.
type Request struct {
	ConnectionID     string         `json:"connection_id,omitempty"`
	ExchangeID       string         `json:"exchange_id,omitempty"`
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
	ClientAddr       string         `json:"client_addr,omitempty"`
	ClientPort       string         `json:"client_port,omitempty"`
	ServerAddr       string         `json:"server_addr,omitempty"`
	ServerPort       string         `json:"server_port,omitempty"`
}

// Kind gets an envelope kind for a Request.
func (r *Request) Kind() envelope.Kind { return "httpx.request" }

// GetConnectionID gets a connection ID.
func (r *Request) GetConnectionID() string { return r.ConnectionID }

// GetExchangeID gets an exchange ID.
func (r *Request) GetExchangeID() string { return r.ExchangeID }

// SetCreated sets the created timestamp
func (r *Request) SetCreated(created time.Time) { r.Created = created }

// SetSessionID sets the session ID
func (r *Request) SetSessionID(id string) { r.SessionID = id }

// StdRequest converts a Request into an *http.Request.
func (r *Request) StdRequest() *http.Request {
	return &http.Request{
		Method:           r.Method,
		URL:              r.URL,
		Proto:            r.Proto,
		ProtoMajor:       r.ProtoMajor,
		ProtoMinor:       r.ProtoMinor,
		Header:           r.Header,
		PostForm:         r.PostForm,
		Body:             ioutil.NopCloser(strings.NewReader(r.Body)),
		TransferEncoding: r.TransferEncoding,
		Host:             r.Host,
		Trailer:          r.Trailer,
		RequestURI:       r.RequestURI,
	}
}

// NewRequest creates a new Request.
func NewRequest(b *bufio.Reader, connectionID string, exchangeID string, m *core.Meta) (*Request, error) {
	req, err := http.ReadRequest(b)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
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
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	var (
		clientAddr string
		clientPort string
		serverAddr string
		serverPort string
		remoteAddr string
	)

	if m != nil {
		var okaddr, okport bool
		clientAddr, okaddr = m.GetString(tcp.MetaSrcAddr)
		clientPort, okport = m.GetString(tcp.MetaSrcPort)

		if okaddr && okport {
			remoteAddr = fmt.Sprintf("%s:%s", clientAddr, clientPort)
		} else {
			remoteAddr = req.RemoteAddr
		}

		serverAddr, _ = m.GetString(tcp.MetaDstAddr)
		serverPort, _ = m.GetString(tcp.MetaDstPort)
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
		RemoteAddr:       remoteAddr,
		RequestURI:       req.RequestURI,
		ClientAddr:       clientAddr,
		ClientPort:       clientPort,
		ServerAddr:       serverAddr,
		ServerPort:       serverPort,
	}, nil
}
