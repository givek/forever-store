package p2p

import "net"

// Message represents any arbitarary data that is being sent over the each
// transport between two nodes in the network.
type Message struct {
	From    net.Addr
	Payload []byte
}
