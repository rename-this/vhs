package pipe

import (
	"context"
	"testing"
	"time"

	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/testhelper"
	"gotest.tools/assert"
)

func TestOutput(t *testing.T) {
	cases := []struct {
		desc string
		o    *Output
		data []interface{}
		out  string
	}{
		{
			desc: "unbuffered",
			o:    NewOutput(testhelper.NewFormat(false), &testhelper.Sink{}),
			data: []interface{}{1, 2, 3},
			out:  `123`,
		},
		{
			desc: "buffered",
			o:    NewOutput(testhelper.NewFormat(true), &testhelper.Sink{}),
			data: []interface{}{1, 2, 3},
			out:  `6`,
		},
		{
			desc: "modifiers",
			o:    NewOutput(testhelper.NewFormat(false), &testhelper.Sink{}, &testhelper.DoubleOutput{}, &testhelper.DoubleOutput{}),
			data: []interface{}{1, 2, 3},
			out:  "111122223333",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			stdCtx, cancel := context.WithCancel(context.Background())
			ctx := &session.Context{StdContext: stdCtx}

			go c.o.Init(ctx)

			for _, d := range c.data {
				c.o.Write(d)
			}

			time.Sleep(500 * time.Millisecond)

			cancel()

			time.Sleep(100 * time.Millisecond)

			s := c.o.Sink.(*testhelper.Sink)
			assert.DeepEqual(t, string(s.Data()), c.out)
		})
	}
}
