package modifier

import (
	"io"

	"github.com/go-errors/errors"
	"go.uber.org/multierr"
)

// ReadCloser wraps an io.ReadCloser
type ReadCloser interface {
	Wrap(io.ReadCloser) (io.ReadCloser, error)
}

// ReadClosers is a slice of ReadCloser.
type ReadClosers []ReadCloser

// Wrap wraps an io.ReadCloser in all Inputs.
func (is ReadClosers) Wrap(r io.ReadCloser) (io.ReadCloser, Closers, error) {
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

// WriteCloser wraps an io.WriteCloser
type WriteCloser interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}

// WriteClosers is a slice of WriteCloser.
type WriteClosers []WriteCloser

// Wrap wraps an io.ReadCloser in all Inputs.
func (os WriteClosers) Wrap(w io.WriteCloser) (io.WriteCloser, Closers, error) {
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
