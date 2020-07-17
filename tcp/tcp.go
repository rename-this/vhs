package tcp

import (
	"time"

	"github.com/google/gopacket/tcpassembly"
)

// BidirectionalStreamFactory is a tcpassembly.StreamFactory
// with additional methods to support bidirectional TCP streams.
type BidirectionalStreamFactory interface {
	tcpassembly.StreamFactory
	Prune(time.Duration)
}
