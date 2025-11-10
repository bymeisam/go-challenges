package main

import (
	"os"
	"testing"
)

func TestGodotenv(t *testing.T) {
	// Create a test .env file
	envContent := `DB_HOST=testhost
DB_PORT=5433
DB_USER=testuser
DB_PASSWORD=testpass
API_KEY=testapikey123
DEBUG=true
`
	
	testFile := ".env.test"
	if err := os.WriteFile(testFile, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(testFile)

	// Load the test .env file
	if err := LoadEnv(testFile); err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}

	// Test environment variables
	if os.Getenv("DB_HOST") != "testhost" {
		t.Errorf("Expected DB_HOST 'testhost', got '%s'", os.Getenv("DB_HOST"))
	}

	if os.Getenv("DB_PORT") != "5433" {
		t.Errorf("Expected DB_PORT '5433', got '%s'", os.Getenv("DB_PORT"))
	}

	// Test GetEnv with default
	value := GetEnv("NON_EXISTENT_KEY", "default_value")
	if value != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%s'", value)
	}

	// Test LoadConfig
	config := LoadConfig()
	if config.DBHost != "testhost" {
		t.Errorf("Expected DBHost 'testhost', got '%s'", config.DBHost)
	}

	if config.DBPassword != "testpass" {
		t.Errorf("Expected DBPassword 'testpass', got '%s'", config.DBPassword)
	}

	if config.APIKey != "testapikey123" {
		t.Errorf("Expected APIKey 'testapikey123', got '%s'", config.APIKey)
	}

	// Clean up environment
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("API_KEY")
	os.Unsetenv("DEBUG")

	t.Log("✓ godotenv works!")
}
