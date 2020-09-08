package flow

import (
	"io"

	"github.com/gramLabs/vhs/internal/ioutilx"
	"go.uber.org/multierr"
)

// InputModifier wraps an io.InputModifier
type InputModifier interface {
	Wrap(ioutilx.ReadCloserID) (ioutilx.ReadCloserID, error)
}

// InputModifiers is a slice of ReadCloser.
type InputModifiers []InputModifier

// Wrap wraps an io.ReadCloser in all Inputs.
func (is InputModifiers) Wrap(r ioutilx.ReadCloserID) (ioutilx.ReadCloserID, ioutilx.Closers, error) {
	closers := make(ioutilx.Closers, 0, len(is))

	var (
		errs error
		err  error
	)

	for i := len(is) - 1; i >= 0; i-- {
		r, err = is[i].Wrap(r)
		errs = multierr.Append(errs, err)
		closers = append(closers, r)
	}

	return r, closers, err
}

// OutputModifier wraps an io.OutputModifier
type OutputModifier interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}

// OutputModifiers is a slice of WriteCloser.
type OutputModifiers []OutputModifier

// Wrap wraps an io.ReadCloser in all Inputs.
func (os OutputModifiers) Wrap(w io.WriteCloser) (io.WriteCloser, ioutilx.Closers, error) {
	closers := make(ioutilx.Closers, 0, len(os))

	var (
		errs error
		err  error
	)

	for i := len(os) - 1; i >= 0; i-- {
		w, err = os[i].Wrap(w)
		errs = multierr.Append(errs, err)
		closers = append(closers, w)
	}

	return w, closers, err
}
