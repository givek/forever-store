package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPPeer struct {
	conn net.Conn

	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr string
	Decoder    Decoder
	ShakeHands HandshakeFunc
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
	ln, err := net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	t.listener = ln

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("startAcceptLoop unexpected error ", err)
			continue
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	err := t.ShakeHands(peer)
	if err != nil {
		fmt.Println("handleConn unexpected error ", err)
		conn.Close()
		return
	}

	fmt.Println("Accepting new connection!", conn, peer)

	msg := &Message{}

	// buf := make([]byte, 2000)

	// Read loop
	for {
		// n, err := conn.Read(buf)

		err := t.Decoder.Decode(conn, msg)
		if err != nil {
			fmt.Println("handleConn unexpected error - read loop ", err)
			continue
		}

		msg.From = conn.LocalAddr()

		// fmt.Printf("Hello Message: %v\n", buf[:n])
		fmt.Printf("Hello Message: %v\n", msg)
	}

}
