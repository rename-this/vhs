package http

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gramLabs/vhs/sink"
)

var _ sink.Sink = &HARSink{}

// HARSink is an HTTP Archive.
// https://w3c.github.io/web-performance/specs/HARSink/Overview.html
// http://www.softwareishard.com/blog/har-12-spec/
type HARSink struct {
	w io.Writer
	c *Correlator

	mu  sync.Mutex
	out har
}

// NewHAR creates a mew HAR sink.
func NewHAR(w io.Writer, reqTimeout time.Duration) *HARSink {
	return &HARSink{
		w: w,
		c: NewCorrelator(reqTimeout),
		out: har{
			Log: harLog{
				Version: "1.2",
				Creator: harCreator{
					Name:    "vhs",
					Version: "0.0.1",
				},
			},
		},
	}
}

// Init initializes the HAR sink.
func (h *HARSink) Init() {
	go h.c.Start()
	go func() {
		for r := range h.c.Exchanges {
			h.addRequest(r)
		}
	}()
}

// Write writes an HTTP message to a HAR.
func (h *HARSink) Write(n interface{}) error {
	switch m := n.(type) {
	case Message:
		h.c.Messages <- m
	}
	return nil
}

func (h *HARSink) addRequest(req *Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

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

	h.out.Log.Entries = append(h.out.Log.Entries, entry)
}

// Flush writes the archive to its underlying writer.
func (h *HARSink) Flush() error {
	if err := json.NewEncoder(h.w).Encode(h.out); err != nil {
		return fmt.Errorf("failed to write HAR: %w", err)
	}
	return nil
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
