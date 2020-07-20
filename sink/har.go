package sink

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var _ Sink = &HAR{}

// HAR is an HTTP Archive.
// https://w3c.github.io/web-performance/specs/HAR/Overview.html
// http://www.softwareishard.com/blog/har-12-spec/
type HAR struct {
	w   io.Writer
	out harOut
}

// NewHAR creates a mew HAR sink.
func NewHAR(w io.Writer) *HAR {
	return &HAR{
		w: w,
		out: harOut{
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

// Write writes an HTTP message to a HAR.
func (h *HAR) Write(n interface{}) error {
	switch n.(type) {
	case *http.Request:
	case *http.Response:
	}
	return nil
}

// Flush writes the archive to its underlying writer.
func (h *HAR) Flush() error {
	if err := json.NewEncoder(h.w).Encode(h.out); err != nil {
		return fmt.Errorf("failed to write HAR: %w", err)
	}
	return nil
}

type harOut struct {
	Log harLog `json:"log"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment"`
}

type harLog struct {
	Version string        `json:"version"`
	Creator harCreator    `json:"creator"`
	Pages   []interface{} `json:"pages"`
	Entries []interface{} `json:"entries"`
	Comment string        `json:"comment"`
}
