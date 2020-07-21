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
		exchangeID int64
		resBuf     bytes.Buffer
		tee        = io.TeeReader(&s.reader, &resBuf)

		reqReader = bufio.NewReader(tee)
		resReader = bufio.NewReader(&resBuf)
	)

	for {
		s.handle(TypeRequest, func() (Message, error) {
			return NewRequest(reqReader, s.conn.ID, exchangeID)
		})
		s.handle(TypeResponse, func() (Message, error) {
			return NewResponse(resReader, s.conn.ID, exchangeID)
		})

		s.conn.LastActivity = time.Now()
		exchangeID++
	}
}

func (s *Stream) handle(t MessageType, parseMessage func() (Message, error)) {
	msg, err := parseMessage()
	if err != nil {
		// TODO(andrewhare): ultraverbose logging.
		return
	}

	// By default, msgOut is the original message.
	// If middleware is defined, this will be
	// overwritten by the middleware output.
	var msgOut = msg

	if s.middleware != nil {
		msgOut, err = s.middleware.ExecMessage(t, msg)
		if err != nil {
			// TODO(andrewhare): get these errors to a logger
			log.Println("failed to run middleware:", err)
			return
		}
	}

	for _, s := range s.sinks {
		if err := s.Write(msgOut); err != nil {
			// TODO(andrewhare): get these errors to a logger
			log.Println("failed to write sink:", err)
		}
	}
}
