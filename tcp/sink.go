package tcp

import (
	"fmt"
	"net"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewSink creates a new TCP sink.
func NewSink(ctx session.Context) (flow.Sink, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx.StdContext, "tcp", ctx.FlowConfig.AddrSink)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP: %w", err)
	}
	return conn, nil
}
