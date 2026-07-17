package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	payload := "Foo not Bar"
	src := bytes.NewReader([]byte(payload))
	dst := new(bytes.Buffer)

	key := newEncryptionKey()

	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(dst.String())

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != len(payload)+16 {
		t.Errorf("nw != len payload + 16")
	}

	if out.String() != payload {
		t.Errorf("out (%v) != payload (%v)\n", out.String(), payload)
	}

	fmt.Println(out.String())
}
