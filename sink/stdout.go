package sink

import (
	"encoding/json"
	"fmt"
	"os"
)

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

func (s *Stdout) Write(n interface{}) error {
	if err := s.encoder.Encode(n); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}
	return nil
}
