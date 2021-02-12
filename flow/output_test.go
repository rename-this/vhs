package flow

import (
	"errors"
	"testing"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/coretest"
	"gotest.tools/v3/assert"
)

func TestOutput(t *testing.T) {
	ctx := core.NewContext(&core.Config{}, &core.FlowConfig{}, nil)
	ctxBuffered := core.NewContext(&core.Config{}, &core.FlowConfig{BufferOutput: true}, nil)

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
				NewOutput(coretest.NewTestOutputFormatNoErr(ctx), nil, &coretest.TestSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  `123`,
		},
		{
			desc: "buffered",
			oo: Outputs{
				NewOutput(coretest.NewTestOutputFormatNoErr(ctxBuffered), nil, &coretest.TestSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  `6`,
		},
		{
			desc: "modifiers",
			oo: Outputs{
				NewOutput(coretest.NewTestOutputFormatNoErr(ctx), core.OutputModifiers{
					&coretest.TestDoubleOutputModifier{},
					&coretest.TestDoubleOutputModifier{},
				}, &coretest.TestSink{}),
			},
			data: []interface{}{1, 2, 3},
			out:  "111122223333",
		},
		{
			desc: "bad modifier",
			oo: Outputs{
				NewOutput(coretest.NewTestOutputFormatNoErr(ctx), core.OutputModifiers{
					&coretest.TestErrOutputModifier{Err: errors.New("111")},
				}, &coretest.TestSink{}),
			},
			errContains: "111",
		},
		{
			desc: "bad modifier closer",
			oo: Outputs{
				NewOutput(coretest.NewTestOutputFormatNoErr(ctx), core.OutputModifiers{
					&coretest.TestDoubleOutputModifier{OptCloseErr: errors.New("111")},
				}, &coretest.TestSink{}),
			},
			errContains: "111",
		},
		{
			desc: "bad sink closer",
			oo: Outputs{
				NewOutput(coretest.NewTestOutputFormatNoErr(ctx), core.OutputModifiers{
					&coretest.TestDoubleOutputModifier{},
				}, &coretest.TestSink{OptCloseErr: errors.New("111")}),
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// hack: Make this big enough to handle any errors we
			// might end up with.
			errs := make(chan error, 10)
			ctx := core.NewContext(nil, nil, errs)

			for _, o := range c.oo {
				go o.Init(ctx)
			}

			for _, d := range c.data {
				for _, o := range c.oo {
					o.Format.In() <- d
				}
			}

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			time.Sleep(500 * time.Millisecond)

			if c.errContains == "" {
				s := c.oo[0].Sink.(*coretest.TestSink)
				assert.DeepEqual(t, string(s.Data()), c.out)
			} else {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
			}
		})
	}
}
