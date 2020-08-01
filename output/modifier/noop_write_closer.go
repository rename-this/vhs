package modifier

import "io"

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return nopCloser{Writer: w}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
