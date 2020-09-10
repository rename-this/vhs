package prunemap

import (
	"context"
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

	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc

	mu         sync.RWMutex
	items      map[string]*item
	evictCache []interface{}
}

// New creates a new Map
func New(itemTTL, pruneInterval time.Duration) *Map {

	m := &Map{
		Evictions: make(chan interface{}),
		ticker:    time.NewTicker(pruneInterval),
		items:     make(map[string]*item),
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-m.ctx.Done():
				return
			case now := <-m.ticker.C:
				m.mu.Lock()
				for k, i := range m.items {
					if now.Sub(i.last) > itemTTL {
						m.evictCache = append(m.evictCache, m.items[k].n)
						delete(m.items, k)
					}
				}
				m.mu.Unlock()

				for len(m.evictCache) > 0 {
					m.Evictions <- m.evictCache[0]
					m.evictCache[0] = nil
					m.evictCache = m.evictCache[1:]
				}
			}
		}
	}()

	return m
}

// Add adds an item to the map or updates the timestamp if the
// item already exists.
func (m *Map) Add(k string, n interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	i, ok := m.items[k]
	if !ok {
		i = &item{n: n}
		m.items[k] = i
	}

	i.last = time.Now()
}

// Remove removes an item.
func (m *Map) Remove(k string) {
	m.mu.Lock()
	delete(m.items, k)
	m.mu.Unlock()
}

// Get gets a value from the map.
func (m *Map) Get(k string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if i, ok := m.items[k]; ok {
		i.last = time.Now()
		return i.n
	}

	return nil
}

// Close closes the map.
func (m *Map) Close() {
	m.ticker.Stop()
	m.cancel()
}
