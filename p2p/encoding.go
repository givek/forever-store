package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (_ GOBDecoder) Decode(r io.Reader, v *RPC) error {
	return gob.NewDecoder(r).Decode(v)
}

type DefaultDecoder struct{}

func (_ DefaultDecoder) Decode(r io.Reader, msg *RPC) error {

	peekBuf := make([]byte, 1)

	_, err := r.Read(peekBuf)
	if err != nil {
		return err // TODO:
	}

	stream := peekBuf[0] == IncomingStream

	// In case of a stream we are not decoding what is being sent over
	// the network.
	// We are just setting Stream true so we can handle that is our logic.
	if stream {
		msg.Stream = true

		// The stream is going on, so we don't need another Read.
		return nil
	}

	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
