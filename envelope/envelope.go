package envelope

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Kind is the name of a type for an envelope.
type Kind string

// Kindify gets the kind of a thing.
type Kindify interface {
	Kind() Kind
}

// Envelope is a wrapper type for serialization
// that preserves the original type information.
type Envelope struct {
	Kind Kind        `json:"kind"`
	Data interface{} `json:"data"`
}

// New creates a new envelope that wraps n.
func New(n Kindify) Envelope {
	return Envelope{
		Kind: n.Kind(),
		Data: n,
	}
}

// Ctor is a constructor that creates a Kindify.
type Ctor func() Kindify

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		ctors: make(map[Kind]Ctor),
	}
}

// Registry holds a map of names to constructors.
type Registry struct {
	mu    sync.RWMutex
	ctors map[Kind]Ctor
}

// Register registers a namer type with the registry.
func (r *Registry) Register(ctor Ctor) {
	r.mu.Lock()
	r.ctors[ctor().Kind()] = ctor
	r.mu.Unlock()
}

// DecodeJSON decodes a stream of JSON.
// Learned this trick from https://eagain.net/articles/go-json-kind/.
func (r *Registry) DecodeJSON(dec *json.Decoder) (interface{}, error) {
	var msg json.RawMessage
	e := Envelope{
		Data: &msg,
	}

	if err := dec.Decode(&e); err != nil {
		return nil, fmt.Errorf("failed to decode envelope: %w", err)
	}

	n, err := r.get(e.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to get value from envelope registry: %w", err)
	}

	if err := json.Unmarshal(msg, n); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON envelope data: %w", err)
	}

	return n, nil
}

func (r *Registry) get(kind Kind) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ctor, ok := r.ctors[kind]
	if !ok {
		return nil, fmt.Errorf("kind not found in envelope registry: %s", kind)
	}
	return ctor(), nil
}
