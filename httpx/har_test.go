package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gramLabs/vhs/session"
	"gotest.tools/v3/assert"
)

func TestHAR(t *testing.T) {
	cases := []struct {
		desc     string
		messages []Message
		out      *har
	}{
		{
			desc: "no messages",
			out: &har{
				Log: harLog{
					Version: "1.2",
					Creator: harCreator{
						Name:    "vhs",
						Version: "0.0.1",
					},
				},
			},
		},
		{
			desc: "basic",
			messages: []Message{
				&Request{
					ConnectionID: "111",
					ExchangeID:   0,
					Body:         "0",
					Header: http.Header{
						"a": []string{"a"},
					},
					URL: newURL("http://example.org"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   0,
					StatusCode:   http.StatusOK,
				},
			},
			out: &har{
				Log: harLog{
					Version: "1.2",
					Creator: harCreator{
						Name:    "vhs",
						Version: "0.0.1",
					},
					Entries: []harEntry{
						{
							StartedDateTime: "0001-01-01T00:00:00Z",
							Request: harRequest{
								URL:      "http://example.org",
								Headers:  []harNVP{{Name: "a", Value: "a"}},
								BodySize: 1,
							},
							Response: harResponse{Status: http.StatusOK},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			h := NewHAR(30 * time.Second)

			var buf safeBuffer
			stdCtx, cancel := context.WithCancel(context.Background())
			ctx := &session.Context{StdContext: stdCtx}

			go h.Init(ctx, &buf)

			for _, m := range c.messages {
				h.In() <- m
				time.Sleep(100 * time.Millisecond)
			}

			cancel()

			time.Sleep(100 * time.Millisecond)

			b2, err := json.Marshal(c.out)
			assert.NilError(t, err)

			assert.DeepEqual(t, bytes.TrimSpace(buf.Bytes()), b2)
		})
	}
}

type safeBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

func (b *safeBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Read(p)
}

func (b *safeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *safeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Bytes()
}

func newURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
