package main

import (
	"os"
	"testing"
)

func TestOSFilepath(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	if GetEnv("TEST_KEY", "default") != "test_value" {
		t.Error("GetEnv failed")
	}
	
	if GetEnv("NONEXISTENT", "default") != "default" {
		t.Error("GetEnv should return default")
	}
	
	path := JoinPath("a", "b", "c")
	if path != filepath.Join("a", "b", "c") {
		t.Error("JoinPath failed")
	}
	
	if GetExtension("file.txt") != ".txt" {
		t.Error("GetExtension failed")
	}
	
	if GetBasename("/path/to/file.txt") != "file.txt" {
		t.Error("GetBasename failed")
	}
	
	t.Log("✓ os and filepath work!")
}
