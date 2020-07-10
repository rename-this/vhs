package http

import (
	"bufio"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gramLabs/vhs/middleware"
	"github.com/gramLabs/vhs/sink"
)

// StreamFactory is a tcpassembly.StreamFactory
type StreamFactory struct {
	Middleware *middleware.Middleware
	Sinks      []sink.Sink
}

// New creates a new StreamFactory.
func (f *StreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &Stream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		mware:     f.Middleware,
		sinks:     f.Sinks,
	}

	go s.run()

	return &s.r
}

// Stream is an HTTP stream decoder.
type Stream struct {
	net       gopacket.Flow
	transport gopacket.Flow
	r         tcpreader.ReaderStream
	mware     *middleware.Middleware
	sinks     []sink.Sink
}

func (s *Stream) run() {
	buf := bufio.NewReader(&s.r)
	for {
		r, err := NewRequest(buf)
		if err != nil {
			// TODO(andrewhare): get these errors to a logger
			log.Println("failed to parse http request:", err)
			continue
		}

		if r == nil {
			continue
		}

		var (
			// By default, r2 is the original request.
			// If middleware is defined, this will be
			// overwritten by the middleware output.
			r2 interface{} = r
		)

		if s.mware != nil {
			r2, err = s.mware.Exec(r)
			if err != nil {
				// TODO(andrewhare): get these errors to a logger
				log.Println("failed to run middleware:", err)
				continue
			}
		}

		for _, s := range s.sinks {
			if err := s.Write(r2); err != nil {
				// TODO(andrewhare): get these errors to a logger
				log.Println("failed to write sink:", err)
			}
		}
	}
}
