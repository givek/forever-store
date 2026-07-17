package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/givek/forever-store/p2p"
)

type FileServerOpts struct {
	EncKey            []byte
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

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (fs *FileServer) Get(key string) (io.Reader, error) {
	if fs.store.Has(key) {
		_, r, err := fs.store.Read(key)
		return r, err
	}

	fmt.Printf("Don't have any file associated with key: %v, locally. Trying to fetch from the network.\n", key)

	msg := Message{
		Payload: MessageGetFile{Key: key},
	}

	err := fs.broadcast(&msg)
	if err != nil {
		return nil, err
	}

	time.Sleep(500 * time.Millisecond)

	for _, peer := range fs.peers {

		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := fs.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		// fileBuffer := new(bytes.Buffer)
		// n, err := io.Copy(fileBuffer, peer)
		// if err != nil {
		// 	return nil, err
		// }

		fmt.Println("Received bytes over the network: ", n)

		peer.CloseStream()
	}

	_, r, err := fs.store.Read(key)
	return r, err
}

func (fs *FileServer) stream(p *Message) error {
	peers := []io.Writer{}

	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(p)
}

func (fs *FileServer) broadcast(msg *Message) error {
	msgBuf := new(bytes.Buffer)
	err := gob.NewEncoder(msgBuf).Encode(msg)
	if err != nil {
		return err
	}

	// // 2. Bordcast this file to all known peers in the network.
	//
	// p := Payload{Key: key, Data: buf.Bytes()}
	//
	// fs.broadcast(p)

	for _, peer := range fs.peers {
		peer.Send([]byte{p2p.IncomingMessage})

		err = peer.Send(msgBuf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServer) Remove(key string) error {
	return nil
}

func (fs *FileServer) Store(key string, r io.Reader) error {
	// TODO: Check if we can do this with io.ReadSeeker
	// Also what is the diff between this tee vs reseting
	// the read pointer to 0 approach?
	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)

	// 1. Store this file to disk.
	n, err := fs.store.Write(key, tee)
	if err != nil {
		return err
	}

	// - Once the reader is read, at this point it will be empty

	fmt.Println("I wrote some many bytes, but how many? ", n)

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: n + 16,
		},
	}

	err = fs.broadcast(&msg)
	if err != nil {
		return err
	}

	// The message were going too fast and close, the second message was
	// getting dropped.
	time.Sleep(3 * time.Millisecond)

	// payload := []byte("Very big file!")
	for _, peer := range fs.peers {
		// err = peer.Send(payload)
		// if err != nil {
		// 	return err
		// }

		peer.Send([]byte{p2p.IncomingStream})

		n, err := copyEncrypt(fs.EncKey, fileBuffer, peer)
		// n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}

		fmt.Println("Copied bytes ", n)
	}

	return nil

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

func (fs *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	fmt.Printf("Received message from: %v - message data: %v\n", from, msg)

	peer, ok := fs.peers[from] // TODO: I would ideally want a net.Addr here
	if !ok {
		fmt.Println("[handleMessageStoreFile] Just before error", fs.StoreRoot, fs.peers, from)
		return fmt.Errorf("peer not found in the peer list!")
	}
	fmt.Println("HELLLO-DCKX")

	// TODO: Maybe we should return n?
	n, err := fs.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Println("HELLLO-XKCD")

	fmt.Printf("[%v] Written %v bytes to disk\n", fs.Transport.Addr(), n)

	// peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()

	return nil
}

func (fs *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !fs.store.Has(msg.Key) {
		return fmt.Errorf("Could not find file for key: %v.", msg.Key)
	}

	fmt.Println("GETTING SOME FILES!")

	fileSize, r, err := fs.store.Read(msg.Key)
	if err != nil {
		return err
	}

	rc, ok := r.(io.ReadCloser)
	if ok {
		defer rc.Close()
	}

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("Could not find from (%v) peer in the peer list.", from)
	}

	peer.Send([]byte{p2p.IncomingStream})

	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("Wriiten %v bytes to %v\n", n, from)

	return nil
}

func (fs *FileServer) handleMessage(from string, msg *Message) error {
	fmt.Printf("received data: %+v\n", msg)
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return fs.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return fs.handleMessageGetFile(from, v)

	}

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
				log.Println("Failed to decode message: ", rpc)
			}

			fmt.Println("HELLO XYZ-PQRS")

			err = fs.handleMessage(rpc.From.String(), &msg)
			if err != nil {
				log.Println("Failed to handle message: ", rpc)
			}

			// fmt.Printf("%+v\n", msg.Payload)
			//
			// peer, ok := fs.peers[rpc.From.String()]
			// if !ok {
			// 	fmt.Println("[FileServer - loop] Just before panic", fs.StoreRoot, fs.peers, rpc.From.String())
			// 	// TODO: Handle gracefully.
			// 	panic("peer not found in the peer list!")
			// }
			//
			// // fmt.Println(peer, string(msg.Payload.([]byte)))
			//
			// buff := make([]byte, 1024)
			// n, err := peer.Read(buff)
			// if err != nil {
			// 	log.Fatal("Failed to read from peer: ", peer.LocalAddr(), err)
			// }
			//
			// n += 1
			//
			// peer.(*p2p.TCPPeer).Wg.Done()

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
	gob.Register(MessageGetFile{})
}
