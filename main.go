package main

import (
	"fmt"
	"log"

	"github.com/givek/forever-store/p2p"
)

func main() {
	fmt.Println("Hello!")

	tr := p2p.NewTCPTransport(":8080")

	err := tr.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
