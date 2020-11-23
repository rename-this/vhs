package httpx

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/session"
)

func TestCorrelator(t *testing.T) {
	timeout := 10 * time.Millisecond

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

	errs := make(chan error)
	ctx := session.NewContexts(&session.Config{DebugHTTPMessages: true}, &session.FlowConfig{}, errs)

	m.Start(ctx)

	// A paired request/response
	m.Messages <- &Request{ConnectionID: "1", ExchangeID: "0"}
	m.Messages <- &Response{ConnectionID: "1", ExchangeID: "0"}

	// Timed-out request
	m.Messages <- &Request{ConnectionID: "2", ExchangeID: "0"}
	time.Sleep(timeout + 100*time.Millisecond)

	ctx.Cancel()

	exchangeMu.RLock()
	assert.Equal(t, 2, exchangeCount)
	exchangeMu.RUnlock()
}

func TestStressCorrelator(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.") // Can be skipped if too time consuming.
	}

	var (
		httptimeout = 1 * time.Second
		testTimeout = 30 * time.Second
		numSenders  = 4

		numExchangesSent = 0
		numTimeoutsSent  = 0
		ExSentMutex      sync.Mutex
		TimeoutSentMutex sync.Mutex

		numExchangesRecv = 0
		numTimeoutsRecv  = 0
		ExRecvMutex      sync.Mutex
		TimeoutRecvMutex sync.Mutex
	)

	var (
		errs    = make(chan error)
		genctx  = session.NewContexts(&session.Config{DebugHTTPMessages: true}, &session.FlowConfig{}, errs)
		recvctx = session.NewContexts(&session.Config{DebugHTTPMessages: true}, &session.FlowConfig{}, errs)
		corctx  = session.NewContexts(&session.Config{DebugHTTPMessages: true}, &session.FlowConfig{}, errs)
	)

	c := NewCorrelator(httptimeout)
	c.Start(corctx)

	for i := 0; i < numSenders; i++ {
		go func() {
			for {
				select {
				case <-genctx.StdContext.Done():
					return
				default:
					connID := strconv.Itoa(rand.Int())
					timeout := rand.Intn(2)

					if timeout == 0 {
						c.Messages <- &Request{ConnectionID: connID, ExchangeID: "0"}
						c.Messages <- &Response{ConnectionID: connID, ExchangeID: "0"}

						ExSentMutex.Lock()
						numExchangesSent++
						ExSentMutex.Unlock()

					} else {
						c.Messages <- &Request{ConnectionID: connID, ExchangeID: "0"}

						TimeoutSentMutex.Lock()
						numTimeoutsSent++
						TimeoutSentMutex.Unlock()
					}
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-recvctx.StdContext.Done():
				return
			case ex := <-c.Exchanges:
				if ex.Response != nil {
					ExRecvMutex.Lock()
					numExchangesRecv++
					ExRecvMutex.Unlock()

				} else {
					TimeoutRecvMutex.Lock()
					numTimeoutsRecv++
					TimeoutRecvMutex.Unlock()
				}
			}
		}
	}()

	time.Sleep(testTimeout)
	genctx.Cancel()
	time.Sleep(6 * httptimeout)
	corctx.Cancel()
	time.Sleep(100 * time.Millisecond)
	recvctx.Cancel()

	ExSentMutex.Lock()
	defer ExSentMutex.Unlock()
	ExRecvMutex.Lock()
	defer ExRecvMutex.Unlock()
	assert.Equal(t, numExchangesSent, numExchangesRecv)

	TimeoutSentMutex.Lock()
	defer TimeoutSentMutex.Unlock()
	TimeoutRecvMutex.Lock()
	defer TimeoutRecvMutex.Unlock()
	assert.Equal(t, numTimeoutsSent, numTimeoutsRecv)
}
