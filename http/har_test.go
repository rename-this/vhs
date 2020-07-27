package http

import (
	"context"
	"io/ioutil"
	_http "net/http"
	"net/url"
	"testing"
	"time"

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
					Header: _http.Header{
						"a": []string{"a"},
					},
					URL: newURL("http://example.org"),
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   0,
					StatusCode:   _http.StatusOK,
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
							Response: harResponse{Status: 200},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			h := NewHAR(ioutil.Discard, 30*time.Second)

			ctx, cancel := context.WithCancel(context.Background())
			go h.Init(ctx)

			for _, m := range c.messages {
				h.In() <- m
				time.Sleep(100 * time.Millisecond)
			}

			cancel()

			assert.DeepEqual(t, <-h.Out(), c.out)
		})
	}
}

func newURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
