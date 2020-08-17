package tcp

func newConn(up *stream) *conn {
	return &conn{
		up: up,
	}
}

type conn struct {
	up       *stream
	down     *stream
	complete bool
}
