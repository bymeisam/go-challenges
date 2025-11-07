package main

import (
	"io"
	"testing"
)

func TestStringReader(t *testing.T) {
	r := &StringReader{data: "hello"}
	buf := make([]byte, 5)
	n, _ := r.Read(buf)
	if n != 5 || string(buf) != "hello" {
		t.Error("StringReader failed")
	}
	t.Log("✓ io.Reader works!")
}

func TestByteWriter(t *testing.T) {
	w := &ByteWriter{}
	w.Write([]byte("test"))
	if string(w.data) != "test" {
		t.Error("ByteWriter failed")
	}
	t.Log("✓ io.Writer works!")
}

func TestCopy(t *testing.T) {
	r := &StringReader{data: "copy me"}
	w := &ByteWriter{}
	io.Copy(w, r)
	if string(w.data) != "copy me" {
		t.Error("io.Copy failed")
	}
	t.Log("✓ io.Copy works!")
}
