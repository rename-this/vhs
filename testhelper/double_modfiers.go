package testhelper

import (
	"io"
	"io/ioutil"

	"github.com/gramLabs/vhs/ioutilx"
)

// DoubleOutput doubles the output stream.
type DoubleOutput struct{}

// Wrap wraps.
func (*DoubleOutput) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	return ioutilx.NopWriteCloser(&doubleOutput{w: w}), nil
}

type doubleOutput struct {
	w io.WriteCloser
}

func (o *doubleOutput) Write(p []byte) (int, error) {
	return o.w.Write(append(p, p...))
}

// DoubleInput doubles the input stream.
type DoubleInput struct{}

// Wrap wraps.
func (*DoubleInput) Wrap(r io.ReadCloser) (io.ReadCloser, error) {
	return ioutil.NopCloser(&doubleInput{r: r}), nil
}

type doubleInput struct {
	r io.ReadCloser
}

func (i *doubleInput) Read(p []byte) (int, error) {
	b, err := ioutil.ReadAll(i.r)
	if err != nil {
		return 0, err
	}
	copy(p, append(b, b...))
	return len(b) * 2, io.EOF
}
