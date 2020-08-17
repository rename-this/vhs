package tcp

import (
	"fmt"

	"github.com/google/gopacket"
)

type streamID struct {
	net       gopacket.Flow
	transport gopacket.Flow
}

func (id *streamID) Reverse() *streamID {
	return &streamID{
		net:       id.net.Reverse(),
		transport: id.transport.Reverse(),
	}
}

func (id *streamID) String() string {
	return fmt.Sprintf("%v:%v", id.net, id.transport)
}
