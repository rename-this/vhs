package http

import (
	"fmt"
	"time"

	"github.com/gramLabs/vhs/internal/prunemap"
)

// Correlator aggregates HTTP requests and responses and
// creates a full exchange once a request's response is recieved.
// Requests that live longer than the timeout without a corresponding
// response are considered as not having a response and returned as-is.
type Correlator struct {
	Requests  chan *Request
	Responses chan *Response
	Exchanges chan *Request

	cache *prunemap.Map
}

// NewCorrelator creates a new correlator.
func NewCorrelator(timeout time.Duration) *Correlator {
	return &Correlator{
		Requests:  make(chan *Request),
		Responses: make(chan *Response),
		Exchanges: make(chan *Request),

		cache: prunemap.New(timeout, timeout*5),
	}
}

// Start starts the correlator.
func (c *Correlator) Start() {
	for {
		select {
		case req := <-c.Requests:
			c.cache.Add(cacheKey(req), req)
		case res := <-c.Responses:
			k := cacheKey(res)
			if req, ok := c.cache.Get(k).(*Request); ok {
				req.Response = res
				c.Exchanges <- req
				c.cache.Remove(k)
			}
		case i := <-c.cache.Evictions:
			if req, ok := i.(*Request); ok {
				c.Exchanges <- req
			}
		}
	}
}

func cacheKey(msg Message) string {
	return fmt.Sprintf("%s/%d", msg.GetConnectionID(), msg.GetExchangeID())
}
