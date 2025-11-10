package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestIO(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)
	
	lines, err := ReadLines(reader)
	if err != nil || len(lines) != 3 {
		t.Error("ReadLines failed")
	}
	
	src := strings.NewReader("test data")
	dst := &bytes.Buffer{}
	
	n, err := CopyData(src, dst)
	if err != nil || n != 9 || dst.String() != "test data" {
		t.Error("CopyData failed")
	}
	
	t.Log("✓ io and bufio work!")
}
