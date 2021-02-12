package core

import (
	"bytes"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

func TestInputModifier(t *testing.T) {
	cases := []struct {
		desc   string
		inputs InputModifiers
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
			inputs: InputModifiers{
				&TestDoubleInputModifier{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			inputs: InputModifiers{
				&TestDoubleInputModifier{},
				&TestDoubleInputModifier{},
				&TestDoubleInputModifier{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			buf := EmptyMeta(ioutil.NopCloser(bytes.NewBufferString(c.in)))
			r, err := c.inputs.Wrap(buf)
			assert.NilError(t, err)

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			assert.Equal(t, c.out, string(b))
		})
	}
}
