package tcp

import "github.com/segmentio/ksuid"

func newConn(up *stream) *conn {
	return &conn{
		id: ksuid.New().String(),
		up: up,
	}
}

type conn struct {
	id       string
	up       *stream
	down     *stream
	complete bool
}
