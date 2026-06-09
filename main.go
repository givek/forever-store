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
		Decoder:    p2p.DefaultDecoder{},
		OnPeer: func(_ p2p.Peer) error {
			// return fmt.Errorf("Failed to register peer.")
			return nil
		},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := tr.Consume()
			fmt.Println(msg)
		}
	}()

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
