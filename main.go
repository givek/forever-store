package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/givek/forever-store/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	fmt.Printf("Creating a new server with ListenAddr: %v and nodes: %v\n", listenAddr, nodes)

	port := listenAddr

	tcpOpts := &p2p.TCPTransportOpts{
		ListenAddr: fmt.Sprintf(":%v", port),
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},

		// TODO: OnPeer
		// OnPeer: func(_ p2p.Peer) error {
		// 	// return fmt.Errorf("Failed to register peer.")
		// 	return nil
		// },
	}

	transport := p2p.NewTCPTransport(tcpOpts)

	fsOpts := FileServerOpts{
		StoreRoot:         fmt.Sprintf("%v-network", port),
		PathTransformFunc: CASPathTranformFunc,
		Transport:         transport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fsOpts)

	transport.OnPeer = s.OnPeer
	// s.Transport.(*p2p.TCPTransport).OnPeer = s.OnPeer

	fmt.Printf("A new server created with ListenAddr: %v and nodes: %v\n", s.StoreRoot, s.peers)

	return s
}

func main() {

	fs1 := makeServer("8080")
	fs2 := makeServer("3000", ":8080")

	go func() {
		err := fs1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(2 * time.Second)

	go func() {
		err := fs2.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(2 * time.Second)

	data := bytes.NewReader([]byte("My big fat data file here!"))

	fs2.StoreData("some-key-june-28", data)

	select {}

}
