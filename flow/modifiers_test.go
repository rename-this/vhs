package flow

import (
	"bytes"
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
func (m *TestDoubleOutputModifier) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	tdom := &testDoubleOutputModifier{w: w}
	if m.optCloseErr == nil {
		return ioutilx.NopWriteCloser(ioutilx.NopWriteCloser(tdom)), nil
	}

	return &errWriteCloser{
		Writer: ioutilx.NopWriteCloser(tdom),
		err:    m.optCloseErr,
	}, nil
}

type testDoubleOutputModifier struct {
	w io.WriteCloser
}

func (o *testDoubleOutputModifier) Write(p []byte) (int, error) {
	return o.w.Write(append(p, p...))
}

type TestErrOutputModifier struct {
	err error
}

func (m *TestErrOutputModifier) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return nil, m.err
}

// TestDoubleInputModifier doubles the input stream.
type TestDoubleInputModifier struct {
	optCloseErr error
}

// Wrap wraps.
func (m *TestDoubleInputModifier) Wrap(r ioutilx.ReadCloserID) (ioutilx.ReadCloserID, error) {
	tdim := &testDoubleInputModifier{r: r}
	if m.optCloseErr == nil {
		return ioutilx.NopReadCloserID(ioutil.NopCloser(tdim)), nil
	}
	return ioutilx.NopReadCloserID(&errReadCloser{
		Reader: ioutil.NopCloser(tdim),
		err:    m.optCloseErr,
	}), nil
}

type testDoubleInputModifier struct {
	r ioutilx.ReadCloserID
}

type TestErrInputModifier struct {
	err error
}

func (m *TestErrInputModifier) Wrap(w ioutilx.ReadCloserID) (ioutilx.ReadCloserID, error) {
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
			buf := ioutilx.NopReadCloserID(ioutil.NopCloser(bytes.NewBufferString(c.in)))
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
