package httpx

import (
	"net/url"
	"time"
)

func newURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}

func makeComparable(m Message) Message {
	switch r := m.(type) {
	case *Request:
		r.ConnectionID = ""
		r.ExchangeID = 0
		r.Created = time.Time{}
	case *Response:
		r.ConnectionID = ""
		r.ExchangeID = 0
		r.Created = time.Time{}
	}
	return m
}
