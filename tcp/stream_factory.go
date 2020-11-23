package tcp

import (
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// Direction is an enum indicating the direction
// of a TCP stream.
type Direction int

const (
	// DirectionUp is the up-stream of a TCP connection.
	DirectionUp Direction = iota
	// DirectionDown is the down-stream of a TCP connection.
	DirectionDown
)

const (
	// MetaDirection is a key for a flow.Meta to retrieve a direction.
	MetaDirection = "tcp.direction"

	MetaSrcAddr = "ip.srcaddr"
	MetaDstAddr = "ip.dstaddr"
	MetaSrcPort = "tcp.srcport"
	MetaDstPort = "tcp.dstport"
)

func newStreamFactory(ctx session.Context, out chan<- flow.InputReader) *streamFactory {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "tcp_stream_factory").
		Logger()

	return &streamFactory{
		ctx:   ctx,
		out:   out,
		conns: make(map[string]*conn),
	}
}

type streamFactory struct {
	ctx session.Context

	outMu  sync.Mutex
	out    chan<- flow.InputReader
	closed bool

	connsMu sync.Mutex
	conns   map[string]*conn
}

type reader struct {
	ctx  session.Context
	rs   tcpreader.ReaderStream
	s    *stream
	meta *flow.Meta
}

func (r *reader) Reassembled(reassembly []tcpassembly.Reassembly) {
	r.rs.Reassembled(reassembly)
	if r.ctx.Config.DebugPackets {
		r.ctx.Logger.Debug().Interface("reassembly", reassembly).Msg("reassembled")
	} else {
		r.ctx.Logger.Debug().Msg("reassembled")
	}
}

func (r *reader) ReassemblyComplete() {
	r.rs.ReassemblyComplete()
	r.s.conn.complete = true
	r.ctx.Logger.Debug().Msg("reassembly complete")
}

func (r *reader) Read(p []byte) (int, error) {
	return r.rs.Read(p)
}

func (r *reader) Close() error {
	return r.rs.Close()
}

func (r *reader) Meta() *flow.Meta {
	return r.meta
}

type stream struct {
	ctx       session.Context
	net       gopacket.Flow
	transport gopacket.Flow
	conn      *conn
}

func (f *streamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	ctx := f.ctx
	ctx.Logger = f.ctx.Logger.With().
		Str("net", net.String()).
		Str("trans", transport.String()).
		Logger()

	ctx.Logger.Debug().Msg("creating new stream")

	s := &stream{
		ctx:       ctx,
		net:       net,
		transport: transport,
	}

	d := f.trackStream(ctx, s)

	r := &reader{
		ctx: ctx,
		rs:  tcpreader.NewReaderStream(),
		s:   s,
		meta: flow.NewMeta(s.conn.id, map[string]interface{}{
			MetaDirection: d,
			MetaSrcAddr:   net.Src().String(),
			MetaSrcPort:   transport.Src().String(),
			MetaDstAddr:   net.Dst().String(),
			MetaDstPort:   transport.Dst().String(),
		}),
	}

	f.outMu.Lock()
	if !f.closed {
		f.out <- r
	}
	f.outMu.Unlock()

	ctx.Logger.Debug().Msg("emitted reader stream")

	return r
}

func (f *streamFactory) Close() {
	f.outMu.Lock()
	f.closed = true
	close(f.out)
	f.outMu.Unlock()
}

// Prune cleans up old streams and connections that are no longer being used.
// This is a stop-the-world garbage collection.
func (f *streamFactory) prune() {
	f.connsMu.Lock()
	defer f.connsMu.Unlock()

	f.ctx.Logger.Debug().Msg("pruning")

	for id, c := range f.conns {
		if c.complete {
			f.ctx.Logger.Debug().Str("id", id).Msg("removing connection")
			delete(f.conns, id)
		}
	}
}

func (f *streamFactory) trackStream(ctx session.Context, s *stream) Direction {
	f.connsMu.Lock()
	defer f.connsMu.Unlock()

	ctx.Logger.Debug().Msg("tracking stream")

	var (
		id = &streamID{net: s.net, transport: s.transport}
		c  = f.conns[id.String()]
		d  = DirectionUp
	)

	if c == nil {
		c = newConn(s)
		f.conns[id.Reverse().String()] = c
		ctx.Logger.Debug().Str("conn_id", c.id).Msg("creating new connection")
	} else {
		c.down = s
		ctx.Logger.Debug().Str("conn_id", c.id).Msg("setting downstream connection")
		d = DirectionDown
	}

	s.conn = c

	return d
}
