package tcp

import (
	"io"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

func newStreamFactory(out chan io.ReadCloser) *streamFactory {
	return &streamFactory{
		out:   out,
		conns: make(map[string]*conn),
	}
}

type streamFactory struct {
	out chan io.ReadCloser

	mu    sync.Mutex
	conns map[string]*conn
}

type reader struct {
	rs tcpreader.ReaderStream
	s  *stream
}

func (r *reader) Reassembled(reassembly []tcpassembly.Reassembly) {
	r.rs.Reassembled(reassembly)
	r.s.conn.complete = true
}

func (r *reader) ReassemblyComplete() {
	r.rs.ReassemblyComplete()
}

type stream struct {
	net       gopacket.Flow
	transport gopacket.Flow
	conn      *conn
}

func (f *streamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &stream{
		net:       net,
		transport: transport,
	}

	f.trackStream(s)

	r := &reader{
		rs: tcpreader.NewReaderStream(),
		s:  s,
	}

	f.out <- &r.rs

	return r
}

// Prune cleans up old streams and connections that are no longer being used.
// This is a stop-the-world garbage collection.
func (f *streamFactory) prune() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for id, c := range f.conns {
		if c.complete {
			delete(f.conns, id)
		}
	}
}

func (f *streamFactory) trackStream(s *stream) {
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
		c.down = s
	}

	s.conn = c
}
