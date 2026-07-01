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

type Message struct {
	Payload any
}

type MessageStoreFile struct{ Key string }

func (s *FileServer) broadcast(p *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{Key: key},
	}

	err := gob.NewEncoder(buf).Encode(msg)
	if err != nil {
		return err
	}

	for _, peer := range fs.peers {
		err = peer.Send(buf.Bytes())
		if err != nil {
			return err
		}
	}

	// time.Sleep(5 * time.Second)
	//
	// payload := []byte("Very big file!")
	// for _, peer := range fs.peers {
	// 	err = peer.Send(payload)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil

	// // TODO: Check if we can do this with io.ReadSeeker
	// // Also what is the diff between this tee vs reseting
	// // the read pointer to 0 approach?
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)
	//
	// // 1. Store this file to disk.
	// err := fs.store.Write(key, tee)
	// if err != nil {
	// 	return err
	// }
	//
	// // - Once the reader is read, at this point it will be empty
	//
	// // 2. Bordcast this file to all known peers in the network.
	//
	// p := Payload{Key: key, Data: buf.Bytes()}
	//
	// fs.broadcast(p)
	//
	// return nil
}

func (fs *FileServer) Stop() {
	close(fs.quitChan)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p

	log.Printf("[%v] LocalAddr: %v -> Peer added: %v\n", fs.StoreRoot, p.LocalAddr().String(), p.RemoteAddr().String())

	return nil
}

func (fs *FileServer) loop() {
	defer func() {
		fs.Transport.Close()
		log.Println("File server stopped due to user quit action.")
	}()

	for {
		select {

		case rpc := <-fs.Transport.Consume():

			msg := Message{}
			err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg)
			if err != nil {
				log.Fatal("Failed to decode message: ", rpc)
			}

			fmt.Printf("%+v\n", msg.Payload)

			peer, ok := fs.peers[rpc.From.String()]
			if !ok {
				fmt.Println("[FileServer - loop] Just before panic", fs.StoreRoot, fs.peers, rpc.From.String())
				// TODO: Handle gracefully.
				panic("peer not found in the peer list!")
			}

			// fmt.Println(peer, string(msg.Payload.([]byte)))

			buff := make([]byte, 1024)
			n, err := peer.Read(buff)
			if err != nil {
				log.Fatal("Failed to read from peer: ", peer.LocalAddr(), err)
			}

			n += 1

			peer.(*p2p.TCPPeer).Wg.Done()

			// fmt.Printf("Received Msg with payload: %v - %v\n", string(msg.Payload.([]byte)), string(buff[:n]))

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

func init() {
	gob.Register(MessageStoreFile{})
}
