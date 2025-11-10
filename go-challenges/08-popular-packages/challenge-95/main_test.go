package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogrus(t *testing.T) {
	log := NewLogger("debug")
	
	if log.Level != logrus.DebugLevel {
		t.Error("Logger level not set correctly")
	}

	// Test structured logging
	var buf bytes.Buffer
	log.SetOutput(&buf)
	
	LogWithFields(log, "info", "Test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	
	output := buf.String()
	if !strings.Contains(output, "Test message") {
		t.Error("Log message not found in output")
	}
	
	// Verify JSON format
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}
	
	if logEntry["key1"] != "value1" {
		t.Error("Structured field not logged correctly")
	}

	t.Log("✓ logrus logging works!")
}
