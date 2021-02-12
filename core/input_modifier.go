package core

import "fmt"

// InputModifier wraps an InputReader.
type InputModifier interface {
	Wrap(InputReader) (InputReader, error)
}

// InputModifiers is a slice of InputModifier.
type InputModifiers []InputModifier

// Wrap wraps an InputReader in all modifiers.
func (is InputModifiers) Wrap(r InputReader) (InputReader, error) {
	var err error
	for i := len(is) - 1; i >= 0; i-- {
		if r, err = is[i].Wrap(r); err != nil {
			return nil, fmt.Errorf("failed to wrap input writer: %w", err)
		}
	}

	return r, nil
}
