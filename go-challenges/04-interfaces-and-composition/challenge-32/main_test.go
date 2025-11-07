package main

import "testing"

func TestBuffer(t *testing.T) {
	var rw ReadWriter = &Buffer{}

	rw.Write("hello")
	if rw.Read() != "hello" {
		t.Error("Buffer should implement ReadWriter")
	}

	t.Log("✓ Interface composition works!")
}
