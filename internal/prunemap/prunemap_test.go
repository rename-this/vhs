package prunemap

import (
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestMap(t *testing.T) {
	var (
		itemTTL       = 10 * time.Millisecond
		pruneDuration = 100 * time.Millisecond
	)

	m := New(itemTTL, pruneDuration)
	defer m.Close()

	var (
		evictionCount int
		evictionMu    sync.RWMutex
	)
	go func() {
		for range m.Evictions {
			evictionMu.Lock()
			evictionCount++
			evictionMu.Unlock()
		}
	}()

	m.Add("a", struct{}{})
	m.Add("b", struct{}{})
	m.Add("c", struct{}{})

	m.mu.RLock()
	assert.Equal(t, 3, len(m.items))
	m.mu.RUnlock()

	time.Sleep(pruneDuration + 10*time.Millisecond)

	m.mu.RLock()
	assert.Equal(t, 0, len(m.items))
	m.mu.RUnlock()

	m.Add("a", "a")
	m.Add("b", "b")
	m.Add("c", "c")

	assert.Equal(t, "a", m.Get("a"))
	assert.Equal(t, "b", m.Get("b"))
	assert.Equal(t, "c", m.Get("c"))

	time.Sleep(pruneDuration + 10*time.Millisecond)

	m.mu.RLock()
	assert.Equal(t, 0, len(m.items))
	m.mu.RUnlock()

	assert.Equal(t, nil, m.Get("a"))

	m.Add("a", struct{}{})
	m.Remove("a")

	evictionMu.RLock()
	assert.Equal(t, 6, evictionCount)
	evictionMu.RUnlock()
}
