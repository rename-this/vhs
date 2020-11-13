package envelope

import (
	"bytes"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

type duck struct {
	Name string
}

func (*duck) Kind() Kind { return "duck" }

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	r.Register(func() Kindify { return &duck{} })

	var (
		d = &duck{Name: "donald"}
		e = New(d)
	)

	b, err := json.Marshal(&e)
	assert.NilError(t, err)

	dec := json.NewDecoder(bytes.NewReader(b))
	d2, err := r.DecodeJSON(dec)
	assert.NilError(t, err)
	assert.DeepEqual(t, d, d2)
}

func TestNewRegistryUnregistered(t *testing.T) {
	r := NewRegistry()

	var (
		d = &duck{Name: "donald"}
		e = New(d)
	)

	b, err := json.Marshal(&e)
	assert.NilError(t, err)

	dec := json.NewDecoder(bytes.NewReader(b))
	_, err = r.DecodeJSON(dec)
	assert.ErrorContains(t, err, "kind not found")
}
