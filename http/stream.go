package http

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gramLabs/vhs/sink"
)

// Stream is an HTTP stream decoder.
type Stream struct {
	id         string
	net        gopacket.Flow
	transport  gopacket.Flow
	reader     tcpreader.ReaderStream
	middleware *Middleware
	sinks      []sink.Sink
	conn       *conn
}

func (s *Stream) run() {
	var (
		transactionID int64
		resBuf        bytes.Buffer
		tee           = io.TeeReader(&s.reader, &resBuf)

		reqReader = bufio.NewReader(tee)
		resReader = bufio.NewReader(&resBuf)
	)

	for {
		s.handle(TypeRequest, func() (interface{}, error) {
			return NewRequest(reqReader, s.conn.ID, transactionID)
		})
		s.handle(TypeResponse, func() (interface{}, error) {
			return NewResponse(resReader, s.conn.ID, transactionID)
		})

		transactionID++
		s.conn.lastActivity = time.Now()
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
