package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {

	opts := TCPTransportOpts{
		ListenAddr:    ":4000",
		HandshakeFunc: NOPHandshakeFunc,
	}

	listenAddr := ":4000"

	tcpTransport := NewTCPTransport(opts)

	assert.Equal(t, listenAddr, tcpTransport.ListenAddr)

	assert.Nil(t, tcpTransport.ListenAndAccept())

}
