package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "some-random-key"
	pathName := CASPathTranformFunc(key)
	fmt.Println(pathName)

	expectedPathName := "b6d18/9bfbd/7dcc6/1ffcd/093ad/b37d4/60f81/f4216"

	if pathName != expectedPathName {
		t.Errorf(
			"Expected: %v :: Got: %v",
			expectedPathName,
			pathName,
		)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{PathTransformFunc: CASPathTranformFunc}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("Some Things IDK."))

	err := s.writeStream("some-test-key-1", data)
	if err != nil {
		t.Error(err)
	}
}
