package httpx

import (
	"bufio"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestNewRequest(t *testing.T) {
	cases := []struct {
		desc        string
		b           *bufio.Reader
		r           *Request
		cID         string
		eID         int64
		errContains string
	}{
		{
			desc:        "EOF",
			b:           bufio.NewReader(strings.NewReader("")),
			errContains: "EOF",
		},
		{
			desc:        "invalid method",
			b:           bufio.NewReader(strings.NewReader(" / HTTP/1.1\r\nheader:foo\r\n\r\n")),
			errContains: "invalid method",
		},
		{
			desc: "success",
			cID:  "111",
			eID:  111,
			b:    bufio.NewReader(strings.NewReader("GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n")),
			r: &Request{
				ConnectionID:  "111",
				ExchangeID:    111,
				Method:        "GET",
				URL:           newURL("/111.html"),
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        http.Header{"Header": {"foo"}},
				Body:          "",
				ContentLength: 0,
				RequestURI:    "/111.html",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			r, err := NewRequest(c.b, c.cID, c.eID)
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
