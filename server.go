package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/givek/forever-store/p2p"
)

type FileServerOpts struct {
	StoreRoot         string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store    *Store
	quitChan chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		PathTransformFunc: opts.PathTransformFunc,
		Root:              opts.StoreRoot,
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitChan:       make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Payload struct {
	Key  string
	Data []byte
}

func (s *FileServer) broadcast(p Payload) error {
	buf := new(bytes.Buffer)

	for _, peer := range s.peers {
		err := gob.NewEncoder(buf).Encode(p)
		if err != nil {
			return err
		}

		err = peer.Send(buf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store this file to disk.
	// 2. Bordcast this file to all known peers in the network.

	return nil
}

func (fs *FileServer) Stop() {
	close(fs.quitChan)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p

	log.Println("Connected with remote peer", p.RemoteAddr())

	return nil
}

func (fs *FileServer) loop() {
	defer func() {
		fs.Transport.Close()
		log.Println("File server stopped due to user quit action.")
	}()

	for {
		select {

		case msg := <-fs.Transport.Consume():
			fmt.Println(msg)

		case <-fs.quitChan:
			return

		}
	}
}

func (fs *FileServer) bootstrapNetwork() error {

	fmt.Println("Bootstraping the network..")

	for _, addr := range fs.BootstrapNodes {

		fmt.Println("Bootstraping the node with addr: ", addr)
		go func(addr string) {
			err := fs.Transport.Dial(addr)
			if err != nil {
				log.Print("[bootstrapNetwork] Failed to Dail", err)
			}
		}(addr)
	}

	return nil
}

func (fs *FileServer) Start() error {
	fmt.Println("[Start] FileServer OnPeer", fs.OnPeer)
	fs.Transport.
		err := fs.Transport.ListenAndAccept()
	if err != nil {
		return err
	}

	if len(fs.BootstrapNodes) != 0 {
		fs.bootstrapNetwork()
	}
	fs.loop()

	return nil
}
