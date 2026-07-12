package p2p

import "net"

const (
	IncomingStream  = 0x2
	IncomingMessage = 0x1
)

// RPC represents any arbitarary data that is being sent over the each
// transport between two nodes in the network.
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
