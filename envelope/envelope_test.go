package envelope

import (
	"testing"
)

type duck struct {
	s string
}

func (d duck) Name() string { return "duck" }

func TestRegistry(t *testing.T) {
}
