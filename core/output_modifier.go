package core

import "fmt"

// OutputModifier wraps an OtputWriter.
type OutputModifier interface {
	Wrap(OutputWriter) (OutputWriter, error)
}

// OutputModifiers is a slice of OutputModifier.
type OutputModifiers []OutputModifier

// Wrap wraps an OutputWrtier in all modifiers.
func (os OutputModifiers) Wrap(w OutputWriter) (OutputWriter, error) {
	var err error
	for i := len(os) - 1; i >= 0; i-- {
		if w, err = os[i].Wrap(w); err != nil {
			return nil, fmt.Errorf("failed to wrap output writer: %w", err)
		}
	}

	return w, err
}
