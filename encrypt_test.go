package main

import (
	"testing"
)

var (
	testText   = "my little pony"
	testSecret string
)

func TestEncrypt(t *testing.T) {
	var err error
	testSecret, err = stringEncrypt(testText)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecrypt(t *testing.T) {
	revealed, err := stringDecrypt(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	if revealed != testText {
		t.Fatal("Decrypted", revealed, "Should be", testText)
	}
}
