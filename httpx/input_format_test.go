package httpx

import (
	"bufio"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/middleware"
	"github.com/rename-this/vhs/session"
	"github.com/rename-this/vhs/tcp"
)

type testMiddleware struct {
	expectedErr error
}

func (m *testMiddleware) Start() error { return nil }
func (m *testMiddleware) Wait() error  { return nil }
func (m *testMiddleware) Close()       {}

func (m *testMiddleware) Exec(_ session.Context, header []byte, n interface{}) (interface{}, error) {
	if m.expectedErr != nil {
		return nil, m.expectedErr
	}
	switch r := n.(type) {
	case *Request:
		r.RequestURI += " 111"
	case *Response:
		r.Status += " 111"
	}
	return n, nil
}

func newTestInputReader(direction tcp.Direction, s string) flow.InputReader {
	var (
		sr = strings.NewReader(s)
		br = bufio.NewReader(sr)
		c  = ioutil.NopCloser(br)
	)
	r := flow.EmptyMeta(c)
	r.Meta().Set(tcp.MetaDirection, direction)
	return r
}

func TestInputFormatInit(t *testing.T) {
	cases := []struct {
		desc        string
		m           middleware.Middleware
		up          flow.InputReader
		down        flow.InputReader
		msgs        []Message
		count       int
		sessionID   string
		errContains string
	}{
		{
			desc: "empty",
			up:   flow.EmptyMeta(ioutil.NopCloser(strings.NewReader(""))),
			down: flow.EmptyMeta(ioutil.NopCloser(strings.NewReader(""))),
		},
		{
			desc:  "no middleware",
			up:    newTestInputReader(tcp.DirectionUp, "GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n"),
			down:  newTestInputReader(tcp.DirectionDown, "HTTP/1.1 204 No Content\r\n\r\n"),
			count: 2,
			msgs: []Message{
				&Request{
					Method:     "GET",
					URL:        newURL("/111.html"),
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header:     http.Header{"Header": {"foo"}},
					MimeType:   "text/plain; charset=utf-8",
					Cookies:    []*http.Cookie{},
					RequestURI: "/111.html",
				},
				&Response{
					Status:     "204 No Content",
					StatusCode: 204,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header:     http.Header{},
					Cookies:    []*http.Cookie{},
				},
			},
		},
		{
			desc:  "middleware",
			up:    newTestInputReader(tcp.DirectionUp, "GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n"),
			down:  newTestInputReader(tcp.DirectionDown, "HTTP/1.1 204 No Content\r\n\r\n"),
			count: 2,
			m:     &testMiddleware{},
			msgs: []Message{
				&Request{
					Method:     "GET",
					URL:        newURL("/111.html"),
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header:     http.Header{"Header": {"foo"}},
					MimeType:   "text/plain; charset=utf-8",
					Cookies:    []*http.Cookie{},
					RequestURI: "/111.html 111",
				},
				&Response{
					Status:     "204 No Content 111",
					StatusCode: 204,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
					Cookies:    []*http.Cookie{},
					Header:     http.Header{},
				},
			},
		},
		{
			desc:  "middleware error",
			up:    newTestInputReader(tcp.DirectionUp, "GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n"),
			down:  newTestInputReader(tcp.DirectionDown, "GET /111.html HTTP/1.1\r\nheader:foo\r\n\r\n"),
			count: 0,
			m: &testMiddleware{
				expectedErr: errors.New("111"),
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			errs := make(chan error, 10)
			ctx := session.NewContexts(&session.Config{Debug: true}, &session.FlowConfig{}, errs)
			ctx.SessionID = c.sessionID
			inputFormat, err := NewInputFormat(ctx)
			assert.NilError(t, err)

			streams := make(chan flow.InputReader)

			go inputFormat.Init(ctx, c.m, streams)

			streams <- c.up
			time.Sleep(50 * time.Millisecond)
			streams <- c.down

			var (
				msgs []Message
				out  = inputFormat.Out()
			)
			for i := 0; i < c.count; i++ {
				m := makeComparable((<-out).(Message))
				msgs = append(msgs, m)
			}

			time.Sleep(50 * time.Millisecond)
			ctx.Cancel()

			if c.errContains == "" {
				assert.DeepEqual(t, msgs, c.msgs)
			} else {
				assert.ErrorContains(t, <-errs, c.errContains)
			}
		})
	}
}
