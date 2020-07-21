package http

import (
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestCorrelator(t *testing.T) {
	var (
		timeout = 10 * time.Millisecond
	)

	m := NewCorrelator(timeout)

	var (
		exchangeCount int
		exchangeMu    sync.RWMutex
	)

	go func() {
		for range m.Exchanges {
			exchangeMu.Lock()
			exchangeCount++
			exchangeMu.Unlock()
		}
	}()

	go m.Start()

	// A paired request/response
	m.Requests <- &Request{ConnectionID: "1", ExchangeID: 0}
	m.Responses <- &Response{ConnectionID: "1", ExchangeID: 0}

	// Timed-out request
	m.Requests <- &Request{ConnectionID: "2", ExchangeID: 0}
	time.Sleep(timeout + 100*time.Millisecond)

	exchangeMu.RLock()
	assert.Equal(t, 2, exchangeCount)
	exchangeMu.RUnlock()
}
