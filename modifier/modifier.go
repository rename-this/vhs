package modifier

import (
	"io"

	"github.com/go-errors/errors"
	"go.uber.org/multierr"
)

// Input wraps an io.Input
type Input interface {
	Wrap(io.ReadCloser) (io.ReadCloser, error)
}

// Inputs is a slice of ReadCloser.
type Inputs []Input

// Wrap wraps an io.ReadCloser in all Inputs.
func (is Inputs) Wrap(r io.ReadCloser) (io.ReadCloser, Closers, error) {
	closers := make(Closers, 0, len(is))

	var (
		errs error
		err  error
	)

	for i := len(is) - 1; i >= 0; i-- {
		r, err = is[i].Wrap(r)
		multierr.Append(errs, err)
		closers = append(closers, r)
	}

	return r, closers, err
}

// Output wraps an io.Output
type Output interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}

// Outputs is a slice of WriteCloser.
type Outputs []Output

// Wrap wraps an io.ReadCloser in all Inputs.
func (os Outputs) Wrap(w io.WriteCloser) (io.WriteCloser, Closers, error) {
	closers := make(Closers, 0, len(os))

	var (
		errs error
		err  error
	)

	for i := len(os) - 1; i >= 0; i-- {
		w, err = os[i].Wrap(w)
		multierr.Append(errs, err)
		closers = append(closers, w)
	}

	return w, closers, err
}

// Closers is a slice of io.Closer
type Closers []io.Closer

// Close closes all closers.
func (cs Closers) Close() error {
	var errs error
	for _, c := range cs {
		if err := c.Close(); err != nil {
			errs = multierr.Append(errs, errors.New(err))
		}
	}
	return errs
}
