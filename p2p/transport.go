package p2p

// Peer is an interface that represent the remote node.
type Peer interface {
}

// Transport is anything that handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, websockets, etc)
type Transport interface {
	ListenAndAccept() error
}
