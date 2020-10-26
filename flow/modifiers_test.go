package flow

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/gramLabs/vhs/internal/ioutilx"
	"gotest.tools/v3/assert"
)

// TestDoubleOutputModifier doubles the output stream.
type TestDoubleOutputModifier struct {
	optCloseErr error
}

// Wrap wraps.
func (m *TestDoubleOutputModifier) Wrap(w OutputWriter) (OutputWriter, error) {
	tdom := &testDoubleOutputModifier{w: w}
	if m.optCloseErr == nil {
		return tdom, nil
	}

	return &errWriteCloser{
		Writer: tdom,
		err:    m.optCloseErr,
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

type TestErrOutputModifier struct {
	err error
}

func (m *TestErrOutputModifier) Wrap(w OutputWriter) (OutputWriter, error) {
	return nil, m.err
}

// TestDoubleInputModifier doubles the input stream.
type TestDoubleInputModifier struct {
	optCloseErr error
}

// Wrap wraps.
func (m *TestDoubleInputModifier) Wrap(r InputReader) (InputReader, error) {
	tdim := &testDoubleInputModifier{r: r}
	if m.optCloseErr == nil {
		return ioutil.NopCloser(tdim), nil
	}
	return &errReadCloser{
		Reader: ioutil.NopCloser(tdim),
		err:    m.optCloseErr,
	}, nil
}

type testDoubleInputModifier struct {
	r InputReader
}

type TestErrInputModifier struct {
	err error
}

func (m *TestErrInputModifier) Wrap(_ InputReader) (InputReader, error) {
	return nil, m.err
}

type errReadCloser struct {
	io.Reader
	err error
}

func (n *errReadCloser) Close() error { return n.err }

type errWriteCloser struct {
	io.Writer
	err error
}

func (n *errWriteCloser) Close() error { return n.err }

func (i *testDoubleInputModifier) Read(p []byte) (int, error) {
	b, err := ioutil.ReadAll(i.r)
	if err != nil {
		return 0, err
	}
	copy(p, append(b, b...))
	return len(b) * 2, io.EOF
}

func TestOutputs(t *testing.T) {
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

func TestInputs(t *testing.T) {
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
			buf := ioutil.NopCloser(bytes.NewBufferString(c.in))
			r, err := c.inputs.Wrap(buf)
			assert.NilError(t, err)

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			assert.Equal(t, c.out, string(b))
		})
	}
}

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
			a:           &TestDoubleOutputModifier{optCloseErr: errors.New("111")},
			b:           &TestDoubleOutputModifier{},
			errContains: "111",
		},
		{
			desc:        "error b",
			a:           &TestDoubleOutputModifier{},
			b:           &TestDoubleOutputModifier{optCloseErr: errors.New("222")},
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
