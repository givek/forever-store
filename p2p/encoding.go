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
	buf := make([]byte, 1028)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
