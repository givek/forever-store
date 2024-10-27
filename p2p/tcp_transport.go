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

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {

	return &TCPTransport{
		TCPTransportOpts: opts,
	}

}

func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	t.listener = listener

	go t.startAcceptLoop()

	return nil

}

func (t *TCPTransport) handleConn(conn net.Conn) {

	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {

		fmt.Println("error occured during handeshake", peer, err)

		conn.Close()

		return

	}

	fmt.Println("New incoming connection", conn, peer)

	// buf := make([]byte, 2000)
	msg := &Message{}
	for {

		// n, err := conn.Read(buf)

		// if err != nil {
		// 	fmt.Println("TCP error: ", err)
		// }		// if err := t.Decoder.Decode(conn, msg); err != nil {
		// 	fmt.Println("TCP error", err)
		// 	continue
		// }

		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Println("TCP error", err)
			continue
		}

		msg.From = conn.RemoteAddr()

		fmt.Println("message: ", msg)

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
