package jsonx

import (
	"testing"
	"time"

	"github.com/rename-this/vhs/internal/safebuffer"
	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

type okapi struct {
	NumStripes int
}

func TestOutputFormat(t *testing.T) {
	cases := []struct {
		desc        string
		buf         *safebuffer.SafeBuffer
		in          []interface{}
		out         string
		errContains string
	}{
		{
			desc: "success",
			buf:  &safebuffer.SafeBuffer{},
			in: []interface{}{
				&okapi{NumStripes: 111},
				&okapi{NumStripes: 222},
				&okapi{NumStripes: 333},
			},
			out: "{\"NumStripes\":111}\n{\"NumStripes\":222}\n{\"NumStripes\":333}\n",
		},
		{
			desc: "bad data",
			buf:  &safebuffer.SafeBuffer{},
			in: []interface{}{
				make(chan struct{}),
			},
			errContains: "unsupported type: chan struct {}",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs      = make(chan error, 1)
				ctx, _, _ = session.NewContexts(nil, errs)
				f, _      = NewOutputFormat(ctx)
			)

			go f.Init(ctx, c.buf)

			for _, n := range c.in {
				f.In() <- n
			}

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			if c.errContains != "" {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}

			assert.Equal(t, c.out, string(c.buf.Bytes()))
		})
	}
}

func TestBufferedOutputFormat(t *testing.T) {
	cases := []struct {
		desc        string
		buf         *safebuffer.SafeBuffer
		in          []interface{}
		out         string
		errContains string
	}{
		{
			desc: "success",
			buf:  &safebuffer.SafeBuffer{},
			in: []interface{}{
				&okapi{NumStripes: 111},
				&okapi{NumStripes: 222},
				&okapi{NumStripes: 333},
			},
			out: "[{\"NumStripes\":111},{\"NumStripes\":222},{\"NumStripes\":333}]\n",
		},
		{
			desc: "bad data",
			buf:  &safebuffer.SafeBuffer{},
			in: []interface{}{
				make(chan struct{}),
			},
			errContains: "unsupported type: chan struct {}",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs      = make(chan error, 1)
				ctx, _, _ = session.NewContexts(nil, errs)
				f, _      = NewBufferedOutputFormat(ctx)
			)

			go f.Init(ctx, c.buf)

			for _, n := range c.in {
				f.In() <- n
			}

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			time.Sleep(500 * time.Millisecond)

			if c.errContains != "" {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}

			assert.Equal(t, c.out, string(c.buf.Bytes()))
		})
	}
}
