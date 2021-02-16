package tcp

import (
	"fmt"
	"net"

	"github.com/rename-this/vhs/core"
)

// NewSink creates a new TCP sink.
func NewSink(ctx core.Context) (core.Sink, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx.StdContext, "tcp", ctx.FlowConfig.AddrSink)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP: %w", err)
	}
	return conn, nil
}
