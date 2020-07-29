package http

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gramLabs/vhs/output"
	"github.com/gramLabs/vhs/session"
)

// Stream is an HTTP stream decoder.
type Stream struct {
	id         string
	net        gopacket.Flow
	transport  gopacket.Flow
	reader     tcpreader.ReaderStream
	middleware *Middleware
	pipes      output.Pipes
	conn       *conn
	sess       *session.Session
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

	msg.SetCreated(time.Now())
	msg.SetSession(s.sess)

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

	s.pipes.Write(msgOut)
}
