package httpx

import (
	"bufio"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/tcp"
)

func TestNewResponset(t *testing.T) {
	cases := []struct {
		desc        string
		b           *bufio.Reader
		r           *Response
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
			desc:        "malformed",
			b:           bufio.NewReader(strings.NewReader("AICHTEETEEPEE/1.1 200 OK")),
			meta:        nil,
			errContains: "malformed HTTP version",
		},
		{
			desc: "success",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("HTTP/1.1 204 No Content\r\n\r\n")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaDstAddr: "10.10.10.1",
				tcp.MetaDstPort: "2346",
				tcp.MetaSrcAddr: "10.10.10.2",
				tcp.MetaSrcPort: "80",
			}),
			r: &Response{
				ConnectionID:  "111",
				ExchangeID:    "111",
				Status:        "204 No Content",
				StatusCode:    http.StatusNoContent,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        http.Header{},
				Cookies:       []*http.Cookie{},
				Body:          "",
				ContentLength: 0,
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
		{
			desc: "success with a cookie",
			cID:  "111",
			eID:  "111",
			b:    bufio.NewReader(strings.NewReader("HTTP/1.1 204 No Content\r\nLocation: /111.html\r\nSet-Cookie: grault=foo\r\n\r\n")),
			meta: flow.NewMeta("source", map[string]interface{}{
				tcp.MetaDstAddr: "10.10.10.1",
				tcp.MetaDstPort: "2346",
				tcp.MetaSrcAddr: "10.10.10.2",
				tcp.MetaSrcPort: "80",
			}),
			r: &Response{
				ConnectionID: "111",
				ExchangeID:   "111",
				Status:       "204 No Content",
				StatusCode:   http.StatusNoContent,
				Proto:        "HTTP/1.1",
				ProtoMajor:   1,
				ProtoMinor:   1,
				Header: http.Header{
					"Set-Cookie": {"grault=foo"},
					"Location":   {"/111.html"},
				},
				Cookies:       []*http.Cookie{{Name: "grault", Value: "foo", Raw: "grault=foo"}},
				Body:          "",
				ContentLength: 0,
				Location:      "/111.html",
				ClientAddr:    "10.10.10.1",
				ClientPort:    "2346",
				ServerAddr:    "10.10.10.2",
				ServerPort:    "80",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			r, err := NewResponse(c.b, c.cID, c.eID, c.meta)
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
