package http

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/google/uuid"
	"github.com/gramLabs/vhs/sink"
)

// NewStreamFactory creates a new factory.
func NewStreamFactory(middleware *Middleware, sinks []sink.Sink) *StreamFactory {
	return &StreamFactory{
		Middleware: middleware,
		Sinks:      sinks,
		conns:      make(map[string]*conn),
	}
}

// StreamFactory is a tcpassembly.StreamFactory
type StreamFactory struct {
	Middleware *Middleware
	Sinks      []sink.Sink

	mu    sync.Mutex
	conns map[string]*conn
}

// New creates a new Stream.
func (f *StreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &Stream{
		net:        net,
		transport:  transport,
		reader:     tcpreader.NewReaderStream(),
		middleware: f.Middleware,
		sinks:      f.Sinks,
	}

	f.trackStream(s)

	go s.run()

	return &s.reader
}

// Prune cleans up old streams and connections that are no longer being used.
// This is a stop-the-world garbage collection.
func (f *StreamFactory) Prune(timeout time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cutoff := time.Now().Add(-timeout)
	for id, c := range f.conns {
		if c.LastActivity.Before(cutoff) {
			delete(f.conns, id)
		}
	}
}

func (f *StreamFactory) trackStream(s *Stream) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var (
		id = &streamID{net: s.net, transport: s.transport}
		c  = f.conns[id.String()]
	)

	if c == nil {
		c = newConn(s)
		f.conns[id.Reverse().String()] = c
	} else {
		c.Down = s
	}

	s.conn = c
}

type streamID struct {
	net       gopacket.Flow
	transport gopacket.Flow
}

func (id *streamID) Reverse() *streamID {
	return &streamID{
		net:       id.net.Reverse(),
		transport: id.transport.Reverse(),
	}
}

func (id *streamID) String() string {
	return fmt.Sprintf("%v:%v", id.net, id.transport)
}

type conn struct {
	ID           string
	Up           *Stream
	Down         *Stream
	LastActivity time.Time
}

func newConn(up *Stream) *conn {
	return &conn{
		ID: uuid.New().String(),
		Up: up,
	}
}
