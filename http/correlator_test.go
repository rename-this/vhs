package http

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())

	go m.Start(ctx)

	// A paired request/response
	m.Messages <- &Request{ConnectionID: "1", ExchangeID: 0}
	m.Messages <- &Response{ConnectionID: "1", ExchangeID: 0}

	// Timed-out request
	m.Messages <- &Request{ConnectionID: "2", ExchangeID: 0}
	time.Sleep(timeout + 100*time.Millisecond)

	cancel()

	exchangeMu.RLock()
	assert.Equal(t, 2, exchangeCount)
	exchangeMu.RUnlock()
}
