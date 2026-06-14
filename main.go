package main

import (
	"fmt"
	"log"
	"time"

	"github.com/givek/forever-store/p2p"
)

func main() {

	port := "8080"

	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: fmt.Sprintf(":%v", port),
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},

		// TODO: OnPeer
		OnPeer: func(_ p2p.Peer) error {
			// return fmt.Errorf("Failed to register peer.")
			return nil
		},
	}

	transport := p2p.NewTCPTransport(tcpOpts)

	fsOpts := FileServerOpts{
		StoreRoot:         fmt.Sprintf("%v-network", port),
		PathTransformFunc: CASPathTranformFunc,
		Transport:         transport,
	}

	fs := NewFileServer(fsOpts)

	go func() {
		time.Sleep(5 * time.Second)
		fs.Stop()
	}()

	err := fs.Start()
	if err != nil {
		log.Fatal(err)
	}

	// go func() {
	// 	for {
	// 		msg := transport.Consume()
	// 		fmt.Println(msg)
	// 	}
	// }()
	//
	// err := transport.ListenAndAccept()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
