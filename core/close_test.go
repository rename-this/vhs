package core

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/rename-this/vhs/internal/ioutilx"
	"gotest.tools/assert"
)

type TestDoubleOutputModifier struct {
	OptCloseErr error
}

// Wrap wraps.
func (m *TestDoubleOutputModifier) Wrap(w OutputWriter) (OutputWriter, error) {
	tdom := &testDoubleOutputModifier{w: w}
	if m.OptCloseErr == nil {
		return tdom, nil
	}

	return &errWriteCloser{
		Writer: tdom,
		err:    m.OptCloseErr,
	}, nil
}

type testDoubleOutputModifier struct {
	w io.WriteCloser
}

func (o *testDoubleOutputModifier) Write(p []byte) (int, error) {
	return o.w.Write(append(p, p...))
}

func (o *testDoubleOutputModifier) Close() error {
	return o.w.Close()
}

type errWriteCloser struct {
	io.Writer
	err error
}

func (n *errWriteCloser) Close() error { return n.err }

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
