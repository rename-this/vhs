package jsonx

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/rename-this/vhs/envelope"
	"github.com/rename-this/vhs/flow"
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
				errs = make(chan error, 1)
				ctx  = session.NewContexts(nil, nil, errs)
				f, _ = NewOutputFormat(ctx)
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

type goose struct {
	Name string
}

func (*goose) Kind() envelope.Kind { return "goose" }

func TestInputFormatInit(t *testing.T) {
	cases := []struct {
		desc        string
		data        string
		out         []*goose
		errContains string
	}{
		{
			desc: "simple",
			data: `
{"kind":"goose","data":{"Name":"Canada"}}
{"kind":"goose","data":{"Name":"Christmas Dinner"}}
{"kind":"goose","data":{"Name":"Grey"}}
`,
			out: []*goose{
				{Name: "Canada"},
				{Name: "Christmas Dinner"},
				{Name: "Grey"},
			},
		},
		{
			desc:        "bad json",
			data:        `{{{`,
			errContains: "failed to decode",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs = make(chan error, 1)
				ctx  = session.NewContexts(&session.Config{
					Debug: true,
				}, &session.FlowConfig{}, errs)
			)

			ctx.Registry.Register(func() envelope.Kindify { return &goose{} })

			inputFormat, err := NewInputFormat(ctx)
			assert.NilError(t, err)

			streams := make(chan flow.InputReader)

			go inputFormat.Init(ctx, nil, streams)

			streams <- flow.EmptyMeta(ioutil.NopCloser(strings.NewReader(c.data)))

			for i := 0; i < len(c.out); i++ {
				assert.DeepEqual(t, c.out[i], <-inputFormat.Out())
			}

			time.Sleep(50 * time.Millisecond)

			ctx.Cancel()

			time.Sleep(50 * time.Millisecond)

			if c.errContains != "" {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}
		})
	}
}
