package sink

import (
	"context"
	"fmt"
	"io"
	"os"
)

var _ Sink = &Stdout{}

// Stdout is a sink that writes JSON to stdout.
type Stdout struct{}

// Init is a no-op.
func (*Stdout) Init(_ context.Context) {}

// Write writes values to stdout.
func (s *Stdout) Write(n interface{}) error {
	if w, ok := n.(io.WriterTo); ok {
		_, err := w.WriteTo(os.Stdout)
		return err
	}

	os.Stdout.WriteString(fmt.Sprintf("%#v", n))
	return nil
}
