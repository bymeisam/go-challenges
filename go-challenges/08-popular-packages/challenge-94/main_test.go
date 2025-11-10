package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestViperConfig(t *testing.T) {
	v, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	// Test defaults
	if v.GetString("server.host") != "localhost" {
		t.Error("Default server host not set correctly")
	}

	if v.GetInt("server.port") != 8080 {
		t.Error("Default server port not set correctly")
	}

	// Test setting values
	v.Set("database.host", "db.example.com")
	v.Set("database.port", 5432)
	v.Set("debug", true)

	config, err := GetConfig(v)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if config.Database.Host != "db.example.com" {
		t.Errorf("Expected database host 'db.example.com', got '%s'", config.Database.Host)
	}

	if config.Database.Port != 5432 {
		t.Errorf("Expected database port 5432, got %d", config.Database.Port)
	}

	if !config.Debug {
		t.Error("Expected debug to be true")
	}

	t.Log("✓ viper configuration works!")
}
