package tcp

import "github.com/google/uuid"

func newConn(up *stream) *conn {
	return &conn{
		id: uuid.New().String(),
		up: up,
	}
}

type conn struct {
	id       string
	up       *stream
	down     *stream
	complete bool
}
