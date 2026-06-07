package main

import (
	"fmt"
	"log"

	"github.com/givek/forever-store/p2p"
)

func main() {
	fmt.Println("Hello!")

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":8080",
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.GOBDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
