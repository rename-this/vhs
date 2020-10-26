package flow

import (
	"errors"
	"testing"
	"time"

	"github.com/gramLabs/vhs/session"
	"gotest.tools/v3/assert"
)

func TestOutput(t *testing.T) {
	ctx, _, _ := session.NewContexts(&session.Config{}, nil)
	ctxBuffered, _, _ := session.NewContexts(&session.Config{BufferOutput: true}, nil)

	cases := []struct {
		desc        string
		oo          Outputs
		data        []interface{}
		out         string
		errContains string
	}{
		{
			desc: "unbuffered",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctx), nil, &testSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  `123`,
		},
		{
			desc: "buffered",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctxBuffered), nil, &testSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  `6`,
		},
		{
			desc: "modifiers",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctx), OutputModifiers{
					&TestDoubleOutputModifier{},
					&TestDoubleOutputModifier{},
				}, &testSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  "111122223333",
		},
		{
			desc: "bad modifier",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctx), OutputModifiers{
					&TestErrOutputModifier{err: errors.New("111")},
				}, &testSink{}),
			},
			errContains: "111",
		},
		{
			desc: "bad modifier closer",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctx), OutputModifiers{
					&TestDoubleOutputModifier{optCloseErr: errors.New("111")},
				}, &testSink{}),
			},
			errContains: "111",
		},
		{
			desc: "bad sink closer",
			oo: Outputs{
				NewOutput(newTestOutputFormatNoErr(ctx), OutputModifiers{
					&TestDoubleOutputModifier{},
				}, &testSink{optCloseErr: errors.New("111")}),
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// hack: Make this big enough to handle any errors we
			// might end up with.
			errs := make(chan error, 10)
			ctx, _, _ := session.NewContexts(nil, errs)

			go c.oo.Init(ctx)

			for _, d := range c.data {
				c.oo.Write(d)
			}

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			time.Sleep(500 * time.Millisecond)

			if c.errContains == "" {
				s := c.oo[0].Sink.(*testSink)
				assert.DeepEqual(t, string(s.Data()), c.out)
			} else {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
			}
		})
	}
}
