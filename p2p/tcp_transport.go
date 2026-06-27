package p2p

import (
	"errors"
	"fmt"
	"net"
)

type TCPPeer struct {
	conn net.Conn

	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
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
	OnPeer     func(p Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	rpcChan chan RPC
}

func NewTCPTransport(opts *TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: *opts,
		rpcChan:          make(chan RPC),
	}
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Consume implements the Transport interface, which will return a read-only
// channel for reading the incoming messages received from another peer in the
// network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

func (t *TCPTransport) ListenAndAccept() error {
	fmt.Println("[ListenAndAccept] OnPeer ", t.OnPeer)
	ln, err := net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	t.listener = ln

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Println("startAcceptLoop unexpected error ", err)
			continue
		}

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	peer := NewTCPPeer(conn, outbound)

	defer peer.Close()

	err := t.ShakeHands(peer)
	if err != nil {
		fmt.Println("handleConn unexpected error ", err)
		conn.Close()
		return
	}

	fmt.Println("Accepting new connection!", conn, peer)

	msg := RPC{}

	fmt.Println("Hello OnPeer: ", t.OnPeer)
	if t.OnPeer != nil {
		err = t.OnPeer(peer)
		if err != nil {
			fmt.Println("Failed to call OnPeer ", err)
			return
		}
	}

	// buf := make([]byte, 2000)

	// Read loop
	for {
		// n, err := conn.Read(buf)

		err := t.Decoder.Decode(conn, &msg)
		// if err == net.ErrClosed {
		// 	fmt.Println("handleConn ErrClosed - read loop ", err)
		// 	return
		// }
		if err != nil {
			fmt.Println("handleConn unexpected error - read loop ", err)
			// continue
			return
		}

		msg.From = conn.LocalAddr()

		// fmt.Printf("Hello Message: %v\n", buf[:n])
		fmt.Printf("Hello Message: %v\n", msg)

		t.rpcChan <- msg
	}

}
