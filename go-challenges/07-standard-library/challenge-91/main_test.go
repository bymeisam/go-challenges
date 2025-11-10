package main

import (
	"flag"
	"os"
	"testing"
)

func TestFlags(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	
	// Simulate command line args
	os.Args = []string{"cmd", "-host=example.com", "-port=3000", "-debug"}
	
	config := ParseFlags()
	
	if config.Host != "example.com" {
		t.Errorf("Host = %s; want example.com", config.Host)
	}
	
	if config.Port != 3000 {
		t.Errorf("Port = %d; want 3000", config.Port)
	}
	
	if !config.Debug {
		t.Error("Debug should be true")
	}
	
	t.Log("✓ flag package works!")
}
