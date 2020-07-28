package http

import (
	"context"
	"io"
	"time"

	"github.com/gramLabs/vhs/output/format"
)

var _ format.Format = &HAR{}

// HAR is an HTTP Archive.
// https://w3c.github.io/web-performance/specs/HAR/Overview.html
// http://www.softwareishard.com/blog/har-12-spec/
type HAR struct {
	w io.Writer
	c *Correlator

	in  chan interface{}
	out chan interface{}
}

// NewHAR creates a mew HAR sink.
func NewHAR(w io.Writer, reqTimeout time.Duration) *HAR {
	return &HAR{
		w:   w,
		c:   NewCorrelator(reqTimeout),
		in:  make(chan interface{}),
		out: make(chan interface{}),
	}
}

// In returns the input channel.
func (h *HAR) In() chan<- interface{} { return h.in }

// Out returns the output channel.
func (h *HAR) Out() <-chan interface{} { return h.out }

// Init initializes the HAR sink.
func (h *HAR) Init(ctx context.Context) {
	go h.c.Start(ctx)

	hh := &har{
		Log: harLog{
			Version: "1.2",
			Creator: harCreator{
				Name:    "vhs",
				Version: "0.0.1",
			},
		},
	}

	for {
		select {
		case n := <-h.in:
			switch m := n.(type) {
			case Message:
				h.c.Messages <- m
			}
		case r := <-h.c.Exchanges:
			h.addRequest(hh, r)
		case <-ctx.Done():
			h.out <- hh
		}
	}
}

func (h *HAR) addRequest(hh *har, req *Request) {
	var headers []harNVP
	for n, vals := range req.Header {
		for _, v := range vals {
			headers = append(headers, harNVP{Name: n, Value: v})
		}
	}

	var queryString []harNVP
	for n, vals := range req.URL.Query() {
		for _, v := range vals {
			queryString = append(queryString, harNVP{Name: n, Value: v})
		}
	}

	request := harRequest{
		Method:      req.Method,
		URL:         req.URL.String(),
		HTTPVersion: req.Proto,
		Headers:     headers,
		QueryString: queryString,
		BodySize:    len(req.Body),
	}

	var response harResponse
	if req.Response != nil {
		var resHeaders []harNVP
		for n, vals := range req.Response.Header {
			for _, v := range vals {
				resHeaders = append(resHeaders, harNVP{Name: n, Value: v})
			}
		}

		content := harContent{
			Size:     req.Response.ContentLength,
			MimeType: req.Response.Header.Get("Content-Type"),
			Text:     req.Response.Body,
		}

		response = harResponse{
			Status:      req.Response.StatusCode,
			StatusText:  req.Response.Status,
			HTTPVersion: req.Response.Proto,
			Headers:     resHeaders,
			Content:     content,
			BodySize:    len(req.Response.Body),
		}
	}

	entry := harEntry{
		StartedDateTime: req.Created.Format(time.RFC3339),
		Request:         request,
		Response:        response,
	}

	hh.Log.Entries = append(hh.Log.Entries, entry)
}

type har struct {
	Log harLog `json:"log"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Entries []harEntry `json:"entries"`
	Comment string     `json:"comment"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment"`
}

type harEntry struct {
	Pageref         string      `json:"pageref,omitempty"`
	StartedDateTime string      `json:"startedDateTime,omitempty"`
	Time            float32     `json:"time,omitempty"`
	Request         harRequest  `json:"request,omitempty"`
	Response        harResponse `json:"response,omitempty"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	Connection      string      `json:"connection,omitempty"`
	Comment         string      `json:"comment,omitempty"`
}

type harRequest struct {
	Method      string   `json:"method,omitempty"`
	URL         string   `json:"url,omitempty"`
	HTTPVersion string   `json:"httpVersion,omitempty"`
	Headers     []harNVP `json:"headers,omitempty"`
	QueryString []harNVP `json:"queryString,omitempty"`
	HeaderSize  int      `json:"headerSize,omitempty"`
	BodySize    int      `json:"bodySize,omitempty"`
	Comment     string   `json:"comment,omitempty"`
}

type harResponse struct {
	Status      int        `json:"status,omitempty"`
	StatusText  string     `json:"statusText,omitempty"`
	HTTPVersion string     `json:"httpVersion,omitempty"`
	Headers     []harNVP   `json:"headers,omitempty"`
	Content     harContent `json:"content,omitempty"`
	RedirectURL string     `json:"redirectURL,omitempty"`
	HeadersSize int        `json:"headersSize,omitempty"`
	BodySize    int        `json:"bodySize,omitempty"`
	Comment     string     `json:"comment,omitempty"`
}

type harNVP struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

type harContent struct {
	Size        int64  `json:"size"`
	Compression int    `json:"compression,omitempty"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Comment     string `json:"comment,omitempty"`
}
