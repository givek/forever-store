package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {

	listenAddr := ":4000"

	tcpTransport := NewTCPTransport(listenAddr)

	assert.Equal(t, listenAddr, tcpTransport.listenAddress)

	assert.Nil(t, tcpTransport.ListenAndAccept())

}
