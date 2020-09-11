package tcp

import (
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

func newStreamFactory(ctx session.Context, out chan ioutilx.ReadCloserID) *streamFactory {
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
	out chan ioutilx.ReadCloserID

	mu    sync.Mutex
	conns map[string]*conn
}

type reader struct {
	ctx session.Context
	rs  tcpreader.ReaderStream
	s   *stream
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

func (r *reader) ID() string {
	return r.s.conn.id
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

	f.trackStream(ctx, s)

	r := &reader{
		ctx: ctx,
		rs:  tcpreader.NewReaderStream(),
		s:   s,
	}

	f.out <- r

	ctx.Logger.Debug().Msg("emitted reader stream")

	return r
}

// Prune cleans up old streams and connections that are no longer being used.
// This is a stop-the-world garbage collection.
func (f *streamFactory) prune() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.ctx.Logger.Debug().Msg("pruning")

	for id, c := range f.conns {
		if c.complete {
			f.ctx.Logger.Debug().Str("id", id).Msg("removing connection")
			delete(f.conns, id)
		}
	}
}

func (f *streamFactory) trackStream(ctx session.Context, s *stream) {
	f.mu.Lock()
	defer f.mu.Unlock()

	ctx.Logger.Debug().Msg("tracking stream")

	var (
		id = &streamID{net: s.net, transport: s.transport}
		c  = f.conns[id.String()]
	)

	if c == nil {
		c = newConn(s)
		f.conns[id.Reverse().String()] = c
		ctx.Logger.Debug().Str("conn_id", c.id).Msg("creating new connection")
	} else {
		c.down = s
		ctx.Logger.Debug().Str("conn_id", c.id).Msg("setting downstream connection")
	}

	s.conn = c
}
