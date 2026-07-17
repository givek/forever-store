package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "some-random-key"
	pk := CASPathTranformFunc(key)

	expectedPathName := "b6d18/9bfbd/7dcc6/1ffcd/093ad/b37d4/60f81/f4216"

	if pk.PathName != expectedPathName {
		t.Errorf(
			"Expected: %v :: Got: %v",
			expectedPathName,
			pk.PathName,
		)
	}
}

func newStore() *Store {
	opts := StoreOpts{PathTransformFunc: CASPathTranformFunc}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	err := s.Clear()
	if err != nil {
		t.Error(err)
	}
}

func TestStoreDeleteKey(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	key := "some-test-key-2"

	bytesData := []byte("Some Things IDK.")

	data := bytes.NewReader(bytesData)

	_, err := s.writeStream(key, data)
	if err != nil {
		t.Error(err)
	}

	err = s.Delete(key)
	if err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := range 50 {
		key := fmt.Sprintf("some-test-key-%v", i)

		bytesData := []byte("Some Things IDK." + " " + key)

		data := bytes.NewReader(bytesData)

		_, err := s.writeStream(key, data)
		if err != nil {
			t.Error(err)
		}

		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(bytesData, b) {
			t.Errorf(
				"Expected: %v :: Got: %v",
				bytesData,
				b,
			)
		}

		err = s.Delete(key)
		if err != nil {
			t.Error(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf(
				"Deleted file still exists.",
			)
		}
	}
}
