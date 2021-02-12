package flow

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/coretest"
	"github.com/rename-this/vhs/internal/ioutilx"
	"gotest.tools/v3/assert"
)

func TestOutputs(t *testing.T) {
	cases := []struct {
		desc    string
		outputs core.OutputModifiers
		in      string
		out     string
	}{
		{
			desc: "no modifiers",
			in:   "111",
			out:  "111",
		},
		{
			desc: "single modifiers",
			in:   "1",
			outputs: core.OutputModifiers{
				&coretest.TestDoubleOutputModifier{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			outputs: core.OutputModifiers{
				&coretest.TestDoubleOutputModifier{},
				&coretest.TestDoubleOutputModifier{},
				&coretest.TestDoubleOutputModifier{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var buf bytes.Buffer
			w, err := c.outputs.Wrap(ioutilx.NopWriteCloser(&buf))
			assert.NilError(t, err)

			_, err = w.Write([]byte(c.in))
			assert.NilError(t, err)

			assert.Equal(t, c.out, buf.String())
		})
	}
}

func TestInputs(t *testing.T) {
	cases := []struct {
		desc   string
		inputs core.InputModifiers
		in     string
		out    string
	}{
		{
			desc: "no modifiers",
			in:   "111",
			out:  "111",
		},
		{
			desc: "single modifiers",
			in:   "1",
			inputs: core.InputModifiers{
				&coretest.TestDoubleInputModifier{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			inputs: core.InputModifiers{
				&coretest.TestDoubleInputModifier{},
				&coretest.TestDoubleInputModifier{},
				&coretest.TestDoubleInputModifier{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			buf := core.EmptyMeta(ioutil.NopCloser(bytes.NewBufferString(c.in)))
			r, err := c.inputs.Wrap(buf)
			assert.NilError(t, err)

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			assert.Equal(t, c.out, string(b))
		})
	}
}
