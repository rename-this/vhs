package modifier

import (
	"fmt"
	"io"
)

// Modifier modifies formatted data.
type Modifier interface {
	Wrap(io.WriteCloser) io.WriteCloser
}

// Modifiers is a slice of Modifier.
type Modifiers []Modifier

// Wrap wraps w in all modifiers.
func (ms Modifiers) Wrap(w io.WriteCloser) (io.WriteCloser, func() error) {
	closers := make([]func() error, 0, len(ms))

	for i := len(ms) - 1; i >= 0; i-- {
		w = ms[i].Wrap(w)
		closers = append(closers, w.Close)
	}

	closeAll := func() error {
		for _, close := range closers {
			if err := close(); err != nil {
				return fmt.Errorf("failed to close modifier: %w", err)
			}
		}
		return nil
	}

	return w, closeAll
}
