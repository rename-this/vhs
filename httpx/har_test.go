package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/internal/safebuffer"
	"github.com/rename-this/vhs/session"
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
					ExchangeID:   "0",
					Body:         "0",
					Header: http.Header{
						"a": []string{"a"},
					},
					URL:        newURL("http://example.org"),
					ServerAddr: "10.10.10.1",
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
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
							Time:            0,
							Request: harRequest{
								URL:        "http://example.org",
								Headers:    []harNVP{{Name: "a", Value: "a"}},
								BodySize:   1,
								HeaderSize: -1,
							},
							Response: harResponse{
								Status:      http.StatusOK,
								HeadersSize: -1,
							},
							Cache: harCache{},
							Timings: harEntryTiming{
								Send:    1,
								Wait:    1,
								Receive: 1,
							},
							ServerIPAddress: "10.10.10.1",
							Connection:      "111",
						},
					},
				},
			},
		},
		{
			desc: "kitchen sink",
			messages: []Message{
				// GET with a cookie
				&Request{
					ConnectionID: "111",
					ExchangeID:   "0",
					Method:       "GET",
					URL:          newURL("/111.html"),
					Proto:        "HTTP/1.1",
					ProtoMajor:   1,
					ProtoMinor:   1,
					Header:       http.Header{"Cookie": {"quux=corge"}},
					MimeType:     "text/plain; charset=utf-8",
					Cookies: []*http.Cookie{
						{
							Name:  "quux",
							Value: "corge",
						},
					},
					Body:          "",
					ContentLength: 1,
					RequestURI:    "/111.html",
					ServerAddr:    "10.10.10.1",
				},
				&Response{
					ConnectionID: "111",
					ExchangeID:   "0",
					StatusCode:   http.StatusOK,
				},
				// POST url-encoded
				&Request{
					ConnectionID: "112",
					ExchangeID:   "1",
					Method:       "POST",
					URL:          newURL("/111.html"),
					Proto:        "HTTP/1.1",
					ProtoMajor:   1,
					ProtoMinor:   1,
					Header: http.Header{
						"Content-Type": {"application/x-www-form-urlencoded"},
					},
					MimeType: "application/x-www-form-urlencoded",
					PostForm: url.Values{
						"baz": {"qux"},
					},
					Cookies:       []*http.Cookie{},
					Body:          "",
					ContentLength: 15,
					RequestURI:    "/111.html",
					ServerAddr:    "10.10.10.2",
				},
				&Response{
					ConnectionID: "112",
					ExchangeID:   "1",
					StatusCode:   http.StatusOK,
				},
				// POST json (not url-encoded)
				&Request{
					ConnectionID: "113",
					ExchangeID:   "1",
					Method:       "POST",
					URL:          newURL("/111.html"),
					Proto:        "HTTP/1.1",
					ProtoMajor:   1,
					ProtoMinor:   1,
					Header: http.Header{
						"Content-Type": {"application/json"},
					},
					MimeType:      "application/json",
					Cookies:       []*http.Cookie{},
					Body:          "{\"baz\":\"qux\",\"foo\":\"bar\"}",
					ContentLength: 25,
					RequestURI:    "/111.html",
					ServerAddr:    "10.10.10.3",
				},
				&Response{
					ConnectionID: "113",
					ExchangeID:   "1",
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
						{ // GET with a cookie
							StartedDateTime: "0001-01-01T00:00:00Z",
							Time:            0,
							Request: harRequest{
								Method:      "GET",
								URL:         "/111.html",
								HTTPVersion: "HTTP/1.1",
								Cookies:     []harCookie{{Name: "quux", Value: "corge"}},
								Headers:     []harNVP{{Name: "Cookie", Value: "quux=corge"}},
								HeaderSize:  -1,
							},
							Response: harResponse{
								Status:      http.StatusOK,
								HeadersSize: -1,
							},
							Cache: harCache{},
							Timings: harEntryTiming{
								Send:    1,
								Wait:    1,
								Receive: 1,
							},
							ServerIPAddress: "10.10.10.1",
							Connection:      "111",
						},
						{ // POST url-encoded
							StartedDateTime: "0001-01-01T00:00:00Z",
							Time:            0,
							Request: harRequest{
								Method:      "POST",
								URL:         "/111.html",
								HTTPVersion: "HTTP/1.1",
								Headers: []harNVP{
									{Name: "Content-Type", Value: "application/x-www-form-urlencoded"},
								},
								PostData: harPOST{
									MIMEType: "application/x-www-form-urlencoded",
									Params: []harNVP{
										{Name: "baz", Value: "qux"},
									},
								},
								HeaderSize: -1,
							},
							Response: harResponse{
								Status:      http.StatusOK,
								HeadersSize: -1,
							},
							Cache: harCache{},
							Timings: harEntryTiming{
								Send:    1,
								Wait:    1,
								Receive: 1,
							},
							ServerIPAddress: "10.10.10.2",
							Connection:      "112",
						},
						{ // POST JSON (not url-encoded)
							StartedDateTime: "0001-01-01T00:00:00Z",
							Time:            0,
							Request: harRequest{
								Method:      "POST",
								URL:         "/111.html",
								HTTPVersion: "HTTP/1.1",
								Headers: []harNVP{
									{Name: "Content-Type", Value: "application/json"},
								},
								PostData: harPOST{
									MIMEType: "application/json",
									Text:     "{\"baz\":\"qux\",\"foo\":\"bar\"}",
								},
								HeaderSize: -1,
								BodySize:   25,
							},
							Response: harResponse{
								Status:      http.StatusOK,
								HeadersSize: -1,
							},
							Cache: harCache{},
							Timings: harEntryTiming{
								Send:    1,
								Wait:    1,
								Receive: 1,
							},
							ServerIPAddress: "10.10.10.3",
							Connection:      "113",
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ctx := session.NewContexts(&session.Config{},
				&session.FlowConfig{
					HTTPTimeout: 30 * time.Second,
				}, nil)

			h, err := NewHAR(ctx)
			assert.NilError(t, err)

			var buf safebuffer.SafeBuffer

			go h.Init(ctx, &buf)

			for _, m := range c.messages {
				h.In() <- m
				time.Sleep(100 * time.Millisecond)
			}

			ctx.Cancel()

			time.Sleep(100 * time.Millisecond)

			b2, err := json.Marshal(c.out)
			assert.NilError(t, err)

			assert.DeepEqual(t, bytes.TrimSpace(buf.Bytes()), b2)
		})
	}
}

func TestExtractCookies(t *testing.T) {
	cases := []struct {
		desc string
		c    []*http.Cookie
		ref  []harCookie
	}{
		{
			desc: "easy",
			c: []*http.Cookie{
				{Name: "foo", Value: "bar"},
			},
			ref: []harCookie{
				{Name: "foo", Value: "bar"},
			},
		},
		{
			desc: "harder",
			c: []*http.Cookie{
				{Name: "foo", Value: "bar"},
				{Name: "baz", Value: "qux", Expires: refTime},
			},
			ref: []harCookie{
				{Name: "foo", Value: "bar"},
				{Name: "baz", Value: "qux", Expires: refTime.Format(time.RFC3339)},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			hc := extractCookies(c.c)
			sort.SliceStable(hc, func(i, j int) bool { return hc[i].Name < hc[j].Name })
			sort.SliceStable(c.ref, func(i, j int) bool { return c.ref[i].Name < c.ref[j].Name })
			assert.DeepEqual(t, hc, c.ref)
		})
	}
}

func TestExtractPostData(t *testing.T) {
	cases := []struct {
		desc string
		req  *Request
		ref  harPOST
	}{
		{
			desc: "url encoded",
			req: &Request{
				Method:   http.MethodPost,
				MimeType: "application/x-www-form-urlencoded",
				PostForm: url.Values{
					"foo": {"bar"},
					"baz": {"qux"},
				},
			},
			ref: harPOST{
				MIMEType: "application/x-www-form-urlencoded",
				Params: []harNVP{
					{Name: "foo", Value: "bar"},
					{Name: "baz", Value: "qux"},
				},
			},
		},
		{
			desc: "not url encoded",
			req: &Request{
				Method:   http.MethodPost,
				MimeType: "application/json",
				Body:     "{\"baz\":\"qux\",\"foo\":\"bar\"}",
			},
			ref: harPOST{
				MIMEType: "application/json",
				Text:     "{\"baz\":\"qux\",\"foo\":\"bar\"}",
			},
		},
		{
			desc: "not POST",
			req: &Request{
				Method: http.MethodGet,
			},
			ref: harPOST{},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			hp := extractPostData(c.req)
			sort.SliceStable(hp.Params, func(i, j int) bool { return hp.Params[i].Name < hp.Params[j].Name })
			sort.SliceStable(c.ref.Params, func(i, j int) bool { return c.ref.Params[i].Name < c.ref.Params[j].Name })
			assert.DeepEqual(t, hp, c.ref)
		})
	}
}

func TestMapToHarNVP(t *testing.T) {
	cases := []struct {
		desc string
		m    map[string][]string
		ref  []harNVP
	}{
		{
			desc: "default",
			m: map[string][]string{
				"foo": {"bar"},
				"baz": {"qux"},
			},
			ref: []harNVP{
				{Name: "foo", Value: "bar"},
				{Name: "baz", Value: "qux"},
			},
		},
		{
			desc: "multiple values",
			m: map[string][]string{
				"foo": {"bar", "baz", "qux"},
			},
			ref: []harNVP{
				{Name: "foo", Value: "bar"},
				{Name: "foo", Value: "baz"},
				{Name: "foo", Value: "qux"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			h := mapToHarNVP(c.m)
			sort.SliceStable(h, func(i, j int) bool { return h[i].Name < h[j].Name })
			sort.SliceStable(c.ref, func(i, j int) bool { return c.ref[i].Name < c.ref[j].Name })
			assert.DeepEqual(t, h, c.ref)
		})
	}
}
