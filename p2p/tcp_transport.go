package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represent the remote node  over a TCP established connection.
type TCPPeer struct {
	// conn is underlying connection of the peer.
	conn net.Conn

	// if we dial and retrieve a conn => outbound = true
	// if we accept and retrieve a conn => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {

	return &TCPTransport{
		listenAddress: listenAddr,
		shakeHands:    NOPHandshakeFunc,
	}

}

func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.listenAddress)

	if err != nil {
		return err
	}

	t.listener = listener

	go t.startAcceptLoop()

	return nil

}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {

	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil {

		fmt.Println("error occured during handeshake", peer, err)

		conn.Close()

		return

	}

	fmt.Println("New incoming connection", conn, peer)

	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
		}
	}

}

func (t *TCPTransport) startAcceptLoop() {

	for {

		conn, err := t.listener.Accept()

		if err != nil {
			fmt.Println("TCP accept error:", err)
		}

		go t.handleConn(conn)

	}

}
