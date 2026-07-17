package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, keyBuf)
	if err != nil {
		panic(err)
	}
	return keyBuf
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// Read the IV from the given io.Reader
	iv := make([]byte, block.BlockSize())
	_, err = src.Read(iv)
	if err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)

	buf := make([]byte, 32*1024)

	nw := block.BlockSize()

	for {

		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return nw, nil

}

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())

	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return 0, err
	}

	// Prepend the IV to the file.
	_, err = dst.Write(iv)
	if err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)

	buf := make([]byte, 32*1024)

	nw := block.BlockSize()

	for {

		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}

			nw += nn
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, err
		}
	}

	return nw, nil
}
