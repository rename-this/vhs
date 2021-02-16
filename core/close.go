package core

import "io"

// CloseSequentially closes a and if there is no
// error, closes b.
func CloseSequentially(a, b io.Closer) error {
	if err := a.Close(); err != nil {
		return err
	}
	if b != nil {
		return b.Close()
	}
	return nil
}
