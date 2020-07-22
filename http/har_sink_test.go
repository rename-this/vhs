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
		out      har
	}{
		{
			desc: "no messages",
			out: har{
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
			out: har{
				Log: harLog{
					Version: "1.2",
					Creator: harCreator{
						Name:    "vhs",
						Version: "0.0.1",
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			har := NewHAR(ioutil.Discard, 30*time.Second)
			har.Init(context.Background())

			for _, m := range c.messages {
				har.Write(m)
			}

			har.mu.Lock()
			assert.DeepEqual(t, har.out, c.out)
			har.mu.Unlock()
		})
	}
}

func newURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
