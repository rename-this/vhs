package prunemap

import (
	"sync"
	"time"
)

type item struct {
	n    interface{}
	last time.Time
}

// Map is a map that self-prunes items
// based on a TTL.
type Map struct {
	Evictions chan interface{}

	mu sync.RWMutex
	m  map[string]*item
}

// New creates a new Map
func New(itemTTL, pruneInterval time.Duration) *Map {
	m := &Map{
		Evictions: make(chan interface{}),
		m:         make(map[string]*item),
	}

	go func() {
		for now := range time.Tick(pruneInterval) {
			m.mu.Lock()
			for k, i := range m.m {
				if now.Sub(i.last) > itemTTL {
					m.Evictions <- m.m[k].n
					delete(m.m, k)
				}
			}
			m.mu.Unlock()
		}
	}()

	return m
}

// Add adds an item to the map.
func (m *Map) Add(k string, n interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	i, ok := m.m[k]
	if !ok {
		i = &item{n: n}
		m.m[k] = i
	}

	i.last = time.Now()
}

// Remove removes an item.
func (m *Map) Remove(k string) {
	m.mu.Lock()
	delete(m.m, k)
	m.mu.Unlock()
}

// Get gets a value from the map.
func (m *Map) Get(k string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if i, ok := m.m[k]; ok {
		i.last = time.Now()
		return i.n
	}

	return nil
}
