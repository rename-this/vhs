package modifier

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/gramLabs/vhs/ioutilx"
	"github.com/gramLabs/vhs/testhelper"
	"gotest.tools/v3/assert"
)

func TestOutputs(t *testing.T) {
	cases := []struct {
		desc    string
		outputs Outputs
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
			outputs: Outputs{
				&testhelper.DoubleOutput{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			outputs: Outputs{
				&testhelper.DoubleOutput{},
				&testhelper.DoubleOutput{},
				&testhelper.DoubleOutput{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var buf bytes.Buffer
			w, closers, err := c.outputs.Wrap(ioutilx.NopWriteCloser(&buf))
			assert.NilError(t, err)
			assert.Equal(t, len(c.outputs), len(closers))

			_, err = w.Write([]byte(c.in))
			assert.NilError(t, err)

			err = closers.Close()
			assert.NilError(t, err)

			assert.Equal(t, c.out, buf.String())
		})
	}
}

func TestInputs(t *testing.T) {
	cases := []struct {
		desc   string
		inputs Inputs
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
			inputs: Inputs{
				&testhelper.DoubleInput{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			inputs: Inputs{
				&testhelper.DoubleInput{},
				&testhelper.DoubleInput{},
				&testhelper.DoubleInput{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			buf := ioutil.NopCloser(bytes.NewBufferString(c.in))
			r, closers, err := c.inputs.Wrap(buf)
			assert.NilError(t, err)
			assert.Equal(t, len(c.inputs), len(closers))

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			err = closers.Close()
			assert.NilError(t, err)

			assert.Equal(t, c.out, string(b))
		})
	}
}
