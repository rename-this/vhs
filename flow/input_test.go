package flow

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/coretest"
	"gotest.tools/v3/assert"
)

func TestInput(t *testing.T) {
	cases := []struct {
		desc        string
		m           core.Middleware
		data        []core.InputReader
		mis         core.InputModifiers
		out         []interface{}
		errContains string
	}{
		{
			desc: "no modifiers",
			data: []core.InputReader{
				core.EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
			out: []interface{}{1, 2, 3},
		},
		{
			desc: "modifiers",
			data: []core.InputReader{
				core.EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
			mis: core.InputModifiers{
				&coretest.TestDoubleInputModifier{},
			},
			out: []interface{}{1, 2, 3, 1, 2, 3},
		},
		{
			desc: "bad modifier",
			data: []core.InputReader{
				core.EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
			mis: core.InputModifiers{
				&coretest.TestErrInputModifier{Err: errors.New("111")},
			},
			errContains: "111",
		},
		{
			desc: "bad modifier closer",
			data: []core.InputReader{
				core.EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
			mis: core.InputModifiers{
				&coretest.TestDoubleInputModifier{OptCloseErr: errors.New("111")},
			},
			errContains: "111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// hack: Make this big enough to handle any errors we
			// might end up with.
			errs := make(chan error, 10)
			ctx := core.NewContext(&core.Config{}, &core.FlowConfig{}, errs)

			var (
				s    = coretest.NewTestSourceData(c.data)
				f, _ = coretest.NewTestInputFormat(ctx)
				i    = NewInput(s, c.mis, f)
			)

			go i.Init(ctx, c.m)

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			time.Sleep(500 * time.Millisecond)

			if c.errContains == "" {
				out := make([]interface{}, 0, len(c.out))
				for j := 0; j < len(c.out); j++ {
					out = append(out, <-i.Format.Out())
				}
				assert.DeepEqual(t, out, c.out)
			} else {
				assert.Equal(t, len(errs), 1)
				assert.ErrorContains(t, <-errs, c.errContains)
			}
		})
	}
}
