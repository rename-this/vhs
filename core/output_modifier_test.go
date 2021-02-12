package core

import (
	"bytes"
	"testing"

	"github.com/rename-this/vhs/internal/ioutilx"
	"gotest.tools/v3/assert"
)

func TestOutputModifier(t *testing.T) {
	cases := []struct {
		desc    string
		outputs OutputModifiers
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
			outputs: OutputModifiers{
				&TestDoubleOutputModifier{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			outputs: OutputModifiers{
				&TestDoubleOutputModifier{},
				&TestDoubleOutputModifier{},
				&TestDoubleOutputModifier{},
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
