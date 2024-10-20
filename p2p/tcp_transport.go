package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {

	return &TCPTransport{
		listenAddress: listenAddr,
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

func (t *TCPTransport) handleConn(conn net.Conn) {

	fmt.Println("New incoming connection", conn)

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
