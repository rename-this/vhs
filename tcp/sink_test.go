package tcp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

func TestNewSink(t *testing.T) {
	cases := []struct {
		desc        string
		setup       func(*testing.T) (string, func())
		errContains string
	}{
		{
			desc: "bad addr",
			setup: func(_ *testing.T) (string, func()) {
				return "1.1.1.1:1111", func() {}
			},
			errContains: "i/o timeout",
		},
		{
			desc: "success",
			setup: func(t *testing.T) (string, func()) {
				l, err := net.Listen("tcp", ":0")
				assert.NilError(t, err)

				go func() {
					_, err := l.Accept()
					assert.NilError(t, err)
				}()

				return l.Addr().String(), func() {
					l.Close()
				}
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			addr, cleanup := c.setup(t)
			defer cleanup()

			ctx := session.NewContexts(nil, &session.FlowConfig{
				AddrSink: addr,
			}, nil)

			ctx.StdContext, ctx.Cancel = context.WithTimeout(ctx.StdContext, 500*time.Millisecond)

			_, err := NewSink(ctx)
			if c.errContains != "" {
				assert.ErrorContains(t, err, c.errContains)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
