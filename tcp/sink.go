package tcp

import (
	"fmt"
	"net"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewSink creates a new TCP sink.
func NewSink(ctx session.Context) (flow.Sink, error) {
	conn, err := net.Dial("tcp", ctx.FlowConfig.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP: %w", err)
	}
	return conn, nil
}
