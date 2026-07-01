package p2p

import "net"

// Peer is an interface that represents a remote node.
type Peer interface {
	net.Conn
	RemoteAddr() net.Addr
	// Close() error
	Send(b []byte) error
}

// Transport is anything that handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, websockets, ...)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
