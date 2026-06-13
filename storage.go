package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	FileName string
}

func (pk PathKey) FirstPathName() (string, error) {
	paths := strings.Split(pk.PathName, "/")

	if len(paths) <= 0 {
		return "", fmt.Errorf("Invalid PathName: %v", pk.PathName)
	}

	return paths[0], nil
}

func (pk PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", pk.PathName, pk.FileName)
}

var CASPathTranformFunc PathTransformFunc = func(key string) PathKey {
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

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

var DefaultTranformFunc PathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
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

func (s *Store) Delete(key string) error {
	pk := s.PathTransformFunc(key)

	firstPathName, err := pk.FirstPathName()
	if err != nil {
		return err
	}

	return os.RemoveAll(firstPathName)
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pk := s.PathTransformFunc(key)
	return os.Open(pk.FullPath())
}

func (s *Store) Read(key string) (io.Reader, error) {
	rc, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, rc)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pk := s.PathTransformFunc(key)

	// TODO: Maybe add a storage folder and gitignore it.

	err := os.MkdirAll(pk.PathName, os.ModePerm)
	if err != nil {
		return err
	}

	// // fileName := "some-filename"
	// buf := new(bytes.Buffer)
	// io.Copy(buf, r)
	//
	// fileNameBytes := md5.Sum(buf.Bytes())
	// fileName := hex.EncodeToString(fileNameBytes[:])
	//
	// pathWithFileName := pk.PathName + "/" + fileName

	pathWithFileName := pk.FullPath()
	f, err := os.Create(pathWithFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
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
