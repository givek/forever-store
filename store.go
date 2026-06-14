package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

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
	// Root is the folder name of the dir, containing all the files and
	// folders created by the store.
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct{ StoreOpts }

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultTranformFunc
	}
	if len(strings.TrimSpace(opts.Root)) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}
func (s *Store) WithRoot(p string) string { return s.Root + "/" + p }

func (s *Store) Has(key string) bool {
	pk := s.PathTransformFunc(key)

	_, err := os.Stat(s.WithRoot(pk.FullPath()))
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pk := s.PathTransformFunc(key)

	firstPathName, err := pk.FirstPathName()
	if err != nil {
		return err
	}

	return os.RemoveAll(s.WithRoot(firstPathName))
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pk := s.PathTransformFunc(key)
	return os.Open(s.WithRoot(pk.FullPath()))
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

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pk := s.PathTransformFunc(key)

	err := os.MkdirAll(s.WithRoot(pk.PathName), os.ModePerm)
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

	pathWithFileName := s.WithRoot(pk.FullPath())
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
