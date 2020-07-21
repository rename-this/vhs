package sink

import (
	"encoding/json"
	"fmt"
	"os"
)

var _ Sink = &Stdout{}

// Stdout is a sink that writes JSON to stdout.
type Stdout struct {
	encoder *json.Encoder
}

// NewStdout returns a new Stdout.
func NewStdout() *Stdout {
	return &Stdout{
		encoder: json.NewEncoder(os.Stdout),
	}
}

// Write writes values to stdout in JSON format.
func (s *Stdout) Write(n interface{}) error {
	if err := s.encoder.Encode(n); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	return nil
}

// Init is a no-op.
func (*Stdout) Init() {}

// Flush is a no-op.
func (*Stdout) Flush() error { return nil }
