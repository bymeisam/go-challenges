package main

import (
	"testing"

	"go.uber.org/zap"
)

func TestZapLogger(t *testing.T) {
	// Test production logger
	prodLogger, err := NewProductionLogger()
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}
	defer prodLogger.Sync()

	prodLogger.Info("Production log message")

	// Test development logger
	devLogger, err := NewDevelopmentLogger()
	if err != nil {
		t.Fatalf("Failed to create development logger: %v", err)
	}
	defer devLogger.Sync()

	devLogger.Info("Development log message")

	// Test custom logger
	customLogger, err := NewCustomLogger("debug")
	if err != nil {
		t.Fatalf("Failed to create custom logger: %v", err)
	}
	defer customLogger.Sync()

	// Test structured logging
	LogWithFields(customLogger, "User action", map[string]interface{}{
		"user_id": 123,
		"action":  "login",
		"ip":      "192.168.1.1",
	})

	// Test sugar logger
	sugar := customLogger.Sugar()
	sugar.Infow("Sugar log",
		"key1", "value1",
		"key2", 42,
	)

	t.Log("✓ zap logger works!")
}
