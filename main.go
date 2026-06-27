package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/givek/forever-store/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
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

	fmt.Println("Original ", s.OnPeer)

	transport.OnPeer = s.OnPeer
	// s.Transport.(*p2p.TCPTransport).OnPeer = s.OnPeer

	fmt.Println("Original tcpOpts", (*tcpOpts).OnPeer)

	return s
}

func main() {

	fs1 := makeServer("8080")
	fs2 := makeServer("3000", ":8080")

	fmt.Println("Main Original tcpOpts", fs1.Transport.(*p2p.TCPTransport).OnPeer)
	go func() {
		err := fs1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := fs2.Start()
	if err != nil {
		log.Fatal(err)
	}

	data := bytes.NewReader([]byte("My big fat data file here!"))

	s2.StoreFile(data)

}
