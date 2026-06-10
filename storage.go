package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) string

var CASPathTranformFunc PathTransformFunc = func(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)

	for i := range paths {
		from := i * blockSize
		to := (i * blockSize) + blockSize

		paths[i] = hashStr[from:to]
	}

	return strings.Join(paths, "/")
}

var DefaultTranformFunc PathTransformFunc = func(key string) string {
	return key
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type Store struct{ StoreOpts }

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)

	// TODO: Maybe add a storage folder and gitignore it.

	err := os.MkdirAll(pathName, os.ModePerm)
	if err != nil {
		return err
	}

	// fileName := "some-filename"
	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	fileNameBytes := md5.Sum(buf.Bytes())
	fileName := hex.EncodeToString(fileNameBytes[:])

	pathWithFileName := pathName + "/" + fileName

	f, err := os.Create(pathWithFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	log.Printf(
		"Successfully written (%v) bytes to disk: %v",
		n,
		pathWithFileName,
	)

	return nil
}
