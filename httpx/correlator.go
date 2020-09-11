package httpx

import (
	"fmt"
	"time"

	"github.com/gramLabs/vhs/internal/prunemap"
	"github.com/gramLabs/vhs/session"
)

// Correlator aggregates HTTP requests and responses and
// creates a full exchange once a request's response is recieved.
// Requests that live longer than the timeout without a corresponding
// response are considered as not having a response and returned as-is.
type Correlator struct {
	Messages  chan Message
	Exchanges chan *Request

	cache *prunemap.Map
}

// NewCorrelator creates a new correlator.
func NewCorrelator(timeout time.Duration) *Correlator {
	return &Correlator{
		Messages:  make(chan Message),
		Exchanges: make(chan *Request),

		cache: prunemap.New(timeout, timeout*5),
	}
}

// Start starts the correlator.
func (c *Correlator) Start(ctx session.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "correlator").
		Logger()

	ctx.Logger.Debug().Msg("start")

	for {
		select {
		case msg := <-c.Messages:
			k := cacheKey(msg)
			switch r := msg.(type) {
			case *Request:
				c.cache.Add(k, r)
				if ctx.Config.DebugHHTTPMessages {
					ctx.Logger.Debug().Interface("request", r).Msg("received request")
				} else {
					ctx.Logger.Debug().Msg("received request")
				}
			case *Response:
				if req, ok := c.cache.Get(k).(*Request); ok {
					req.Response = r
					c.Exchanges <- req
					c.cache.Remove(k)
					if ctx.Config.DebugHHTTPMessages {
						ctx.Logger.Debug().Interface("response", r).Msg("received response")
					} else {
						ctx.Logger.Debug().Msg("received response")
					}
				}
			}
		case i := <-c.cache.Evictions:
			if req, ok := i.(*Request); ok {
				c.Exchanges <- req
				if ctx.Config.DebugHHTTPMessages {
					ctx.Logger.Debug().Interface("request", req).Msg("evicting request")
				} else {
					ctx.Logger.Debug().Msg("evicting request")
				}
			}
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}

func cacheKey(msg Message) string {
	return fmt.Sprintf("%s/%d", msg.GetConnectionID(), msg.GetExchangeID())
}
