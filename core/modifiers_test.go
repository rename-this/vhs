package core

import (
	"io"
	"io/ioutil"
)

// These are duplicate types from coretest, copied here to
// prevent a import cycle.

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

type TestDoubleInputModifier struct{}

func (m *TestDoubleInputModifier) Wrap(r InputReader) (InputReader, error) {
	return &testDoubleInputModifier{r: r}, nil
}

type testDoubleInputModifier struct {
	r InputReader
}

func (i *testDoubleInputModifier) Read(p []byte) (int, error) {
	b, err := ioutil.ReadAll(i.r)
	if err != nil {
		return 0, err
	}
	copy(p, append(b, b...))
	return len(b) * 2, io.EOF
}

func (*testDoubleInputModifier) Close() error { return nil }
func (*testDoubleInputModifier) Meta() *Meta  { return nil }
