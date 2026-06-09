package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	listenAddr := ":4000"
	opts := TCPTransportOpts{
		ListenAddr: listenAddr,
		Decoder:    DefaultDecoder{},
		ShakeHands: NOPHandshakeFunc,
	}

	tr := NewTCPTransport(opts)

	assert.Equal(t, tr.ListenAddr, listenAddr)
	assert.Nil(t, tr.ListenAndAccept())
}
