package http

import (
	"bufio"
	"bytes"
	"io"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gramLabs/vhs/sink"
)

// StreamFactory is a tcpassembly.StreamFactory
type StreamFactory struct {
	Middleware *Middleware
	Sinks      []sink.Sink
}

// New creates a new StreamFactory.
func (f *StreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &Stream{
		net:        net,
		transport:  transport,
		r:          tcpreader.NewReaderStream(),
		middleware: f.Middleware,
		sinks:      f.Sinks,
	}

	go s.run()

	return &s.r
}

// Stream is an HTTP stream decoder.
type Stream struct {
	net        gopacket.Flow
	transport  gopacket.Flow
	r          tcpreader.ReaderStream
	middleware *Middleware
	sinks      []sink.Sink
}

func (s *Stream) run() {
	var (
		resBuf bytes.Buffer
		tee    = io.TeeReader(&s.r, &resBuf)

		reqReader = bufio.NewReader(tee)
		resReader = bufio.NewReader(&resBuf)
	)

	for {
		s.handle(TypeRequest, func() (interface{}, error) {
			return NewRequest(reqReader)
		})
		s.handle(TypeResponse, func() (interface{}, error) {
			return NewResponse(resReader)
		})
	}
}

func (s *Stream) handle(t MessageType, parseMessage func() (interface{}, error)) {
	r, err := parseMessage()
	if err != nil {
		// TODO(andrewhare): ultraverbose logging.
		return
	}

	// By default, r2 is the original message.
	// If middleware is defined, this will be
	// overwritten by the middleware output.
	var r2 = r

	if s.middleware != nil {
		r2, err = s.middleware.ExecMessage(t, r)
		if err != nil {
			// TODO(andrewhare): get these errors to a logger
			log.Println("failed to run middleware:", err)
			return
		}
	}

	for _, s := range s.sinks {
		if err := s.Write(r2); err != nil {
			// TODO(andrewhare): get these errors to a logger
			log.Println("failed to write sink:", err)
		}
	}
}
