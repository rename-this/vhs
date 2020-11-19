package httpx

import (
	"bufio"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/tcp"
)

func TestNewRequest(t *testing.T) {
	cases := []struct {
		desc        string
		b           *bufio.Reader
		r           *Request
		cID         string
		eID         string
		meta        *flow.Meta
		errContains string
	}{
		{
			desc:        "EOF",
			b:           bufio.NewReader(strings.NewReader("")),
			meta:        nil,
			errContains: "EOF",
		},
		{
			desc:        "invalid method",
			b:           bufio.NewReader(strings.NewReader(" / HTTP/1.1\r\nheader:foo\r\n\r\n")),
			meta:        nil,
			errContains: "invalid method",
		},
		{
			desc: "success",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaSrcAddr: "10.10.10.1",
				tcp.MetaSrcPort: "2346",
				tcp.MetaDstAddr: "10.10.10.2",
				tcp.MetaDstPort: "80",
			}),
			r: &Request{
				ConnectionID:  "111",
				ExchangeID:    "111",
				Method:        "GET",
				URL:           newURL("/111.html"),
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        http.Header{"Header": {"foo"}},
				MimeType:      "text/plain; charset=utf-8",
				Cookies:       []*http.Cookie{},
				Body:          "",
				ContentLength: 0,
				RequestURI:    "/111.html",
				RemoteAddr:    "10.10.10.1:2346",
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
		{
			desc: "cookie",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("GET /111.html HTTP/1.1\r\nCookie: quux=corge\r\n\r\n")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaSrcAddr: "10.10.10.1",
				tcp.MetaSrcPort: "2346",
				tcp.MetaDstAddr: "10.10.10.2",
				tcp.MetaDstPort: "80",
			}),
			r: &Request{
				ConnectionID: "111",
				ExchangeID:   "111",
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
				ContentLength: 0,
				RequestURI:    "/111.html",
				RemoteAddr:    "10.10.10.1:2346",
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
		{
			desc: "post form",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("POST /111.html HTTP/1.1\r\nContent-Length: 15\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nbaz=qux&foo=bar")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaSrcAddr: "10.10.10.1",
				tcp.MetaSrcPort: "2346",
				tcp.MetaDstAddr: "10.10.10.2",
				tcp.MetaDstPort: "80",
			}),
			r: &Request{
				ConnectionID: "111",
				ExchangeID:   "111",
				Method:       "POST",
				URL:          newURL("/111.html"),
				Proto:        "HTTP/1.1",
				ProtoMajor:   1,
				ProtoMinor:   1,
				Header: http.Header{
					"Content-Length": {"15"},
					"Content-Type":   {"application/x-www-form-urlencoded"},
				},
				MimeType: "application/x-www-form-urlencoded",
				PostForm: url.Values{
					"baz": {"qux"},
					"foo": {"bar"},
				},
				Cookies:       []*http.Cookie{},
				Body:          "",
				ContentLength: 15,
				RequestURI:    "/111.html",
				RemoteAddr:    "10.10.10.1:2346",
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
		{
			desc: "post JSON",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("POST /111.html HTTP/1.1\r\nContent-Length: 25\r\nContent-Type: application/json\r\n\r\n{\"baz\":\"qux\",\"foo\":\"bar\"}")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaSrcAddr: "10.10.10.1",
				tcp.MetaSrcPort: "2346",
				tcp.MetaDstAddr: "10.10.10.2",
				tcp.MetaDstPort: "80",
			}),
			r: &Request{
				ConnectionID: "111",
				ExchangeID:   "111",
				Method:       "POST",
				URL:          newURL("/111.html"),
				Proto:        "HTTP/1.1",
				ProtoMajor:   1,
				ProtoMinor:   1,
				Header: http.Header{
					"Content-Length": {"25"},
					"Content-Type":   {"application/json"},
				},
				MimeType:      "application/json",
				Cookies:       []*http.Cookie{},
				Body:          "{\"baz\":\"qux\",\"foo\":\"bar\"}",
				ContentLength: 25,
				RequestURI:    "/111.html",
				RemoteAddr:    "10.10.10.1:2346",
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			r, err := NewRequest(c.b, c.cID, c.eID, c.meta)
			if c.errContains != "" {
				assert.ErrorContains(t, err, c.errContains)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, r, c.r)
				now := time.Now()
				r.SetCreated(now)
				assert.Equal(t, now, r.Created)
				r.SetSessionID("111")
				assert.Equal(t, "111", r.SessionID)
			}
		})
	}
}
