package httpx

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/internal/safebuffer"
	"github.com/segmentio/ksuid"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type badmsg struct{}

func (b *badmsg) GetConnectionID() string {
	panic("not implemented") // TODO: Implement
}

func (b *badmsg) GetExchangeID() string {
	panic("not implemented") // TODO: Implement
}

func (b *badmsg) SetCreated(_ time.Time) {
	panic("not implemented") // TODO: Implement
}

func (b *badmsg) SetSessionID(_ string) {
	panic("not implemented") // TODO: Implement
}

func TestNewOutputFormat(t *testing.T) {
	var (
		start = time.Now().Add(-time.Hour)
		u, _  = url.Parse("http://example.com")
		n     = ksuid.New().String()
		cl    = int64(len(n) + 3)
	)
	cases := []struct {
		desc        string
		messages    []Message
		errContains string
	}{
		{
			desc: "only requests",
			messages: []Message{
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_0\n", Created: start.Add(100 * time.Millisecond)},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_1\n", Created: start.Add(100 * time.Millisecond)},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_2\n", Created: start.Add(100 * time.Millisecond)},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_3\n", Created: start.Add(100 * time.Millisecond)},
			},
		},
		{
			desc: "requests and responses",
			messages: []Message{
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_0\n", Created: start.Add(100 * time.Millisecond)},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_1\n", Created: start.Add(100 * time.Millisecond)},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_2\n", Created: start.Add(100 * time.Millisecond)},
				&Response{},
				&Request{Method: "POST", URL: u, ContentLength: cl, Body: n + "_4\n", Created: start.Add(100 * time.Millisecond)},
				&Response{},
				&Response{},
			},
		},
		{
			desc: "wrong message type",
			messages: []Message{
				&badmsg{},
			},
			errContains: "unknown type",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs = make(chan error, 1)
				ctx  = core.NewContext(nil, nil, errs)
				o, _ = NewOutputFormat(ctx)
				in   = o.In()
				w    safebuffer.SafeBuffer
			)

			go o.Init(ctx, &w)

			for _, m := range c.messages {
				in <- m
			}

			time.Sleep(time.Second)

			ctx.Cancel()

			if c.errContains != "" {
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}

			allReqs := string(w.Bytes())

			for i, r := range c.messages {
				if _, ok := r.(*Request); !ok {
					continue
				}
				nn := fmt.Sprintf("%s_%d", n, i)
				assert.Assert(t, cmp.Contains(allReqs, nn))
			}
		})
	}
}
