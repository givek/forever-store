package main

import (
	"fmt"
	"log"

	"github.com/givek/forever-store/p2p"
)

type FileServerOpts struct {
	StoreRoot         string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
}

type FileServer struct {
	FileServerOpts

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
	}
}

func (fs *FileServer) Stop() {
	close(fs.quitChan)
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

func (fs *FileServer) Start() error {
	err := fs.Transport.ListenAndAccept()
	if err != nil {
		return err
	}

	fs.loop()

	return nil
}
