package core

import (
	"bytes"
	"errors"
	"testing"

	"github.com/rename-this/vhs/internal/ioutilx"
	"gotest.tools/assert"
)

func TestCloseSequentially(t *testing.T) {
	cases := []struct {
		desc        string
		a           OutputModifier
		b           OutputModifier
		errContains string
	}{
		{
			desc: "no error",
			a:    &TestDoubleOutputModifier{},
			b:    &TestDoubleOutputModifier{},
		},
		{
			desc:        "error a",
			a:           &TestDoubleOutputModifier{OptCloseErr: errors.New("111")},
			b:           &TestDoubleOutputModifier{},
			errContains: "111",
		},
		{
			desc:        "error b",
			a:           &TestDoubleOutputModifier{},
			b:           &TestDoubleOutputModifier{OptCloseErr: errors.New("222")},
			errContains: "222",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var bufA bytes.Buffer
			a, err := c.a.Wrap(ioutilx.NopWriteCloser(&bufA))
			assert.NilError(t, err)

			var bufB bytes.Buffer
			b, err := c.b.Wrap(ioutilx.NopWriteCloser(&bufB))
			assert.NilError(t, err)

			err = CloseSequentially(a, b)
			if c.errContains == "" {
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}
