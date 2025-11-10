package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	buf := &bytes.Buffer{}
	
	LogMessage(buf, "test message")
	output := buf.String()
	
	if !strings.Contains(output, "INFO:") || !strings.Contains(output, "test message") {
		t.Errorf("Log output incorrect: %s", output)
	}
	
	buf.Reset()
	LogError(buf, errors.New("test error"))
	output = buf.String()
	
	if !strings.Contains(output, "ERROR:") || !strings.Contains(output, "test error") {
		t.Errorf("Error log incorrect: %s", output)
	}
	
	t.Log("✓ log package works!")
}
