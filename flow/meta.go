package flow

import (
	"io"
	"sync"
)

// EmptyMeta wraps an io.ReadCloser with an empty Meta.
func EmptyMeta(r io.ReadCloser) InputReader {
	return &emptyMeta{
		ReadCloser: r,
		meta:       NewMeta("", nil),
	}
}

type emptyMeta struct {
	io.ReadCloser
	meta *Meta
}

func (m *emptyMeta) Meta() *Meta {
	return m.meta
}

// NewMeta creates a new Meta.
func NewMeta(sourceID string, values map[string]interface{}) *Meta {
	if values == nil {
		values = make(map[string]interface{})
	}

	return &Meta{
		SourceID: sourceID,
		values:   values,
	}
}

// Meta is a map of flow-related metadata.
type Meta struct {
	SourceID string

	mu     sync.RWMutex
	values map[string]interface{}
}

// Set sets a value.
func (m *Meta) Set(key string, value interface{}) {
	m.mu.Lock()
	m.values[key] = value
	m.mu.Unlock()
}

// Get gets a value.
func (m *Meta) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.values[key]
	return v, ok
}

// GetString gets a string value.
func (m *Meta) GetString(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.values[key].(string)
	return v, ok
}
