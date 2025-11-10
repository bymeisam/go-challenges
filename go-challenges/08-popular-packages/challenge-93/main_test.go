package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCobraCLI(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewGreetCmd())
	rootCmd.AddCommand(NewVersionCmd())

	// Test root command exists
	if rootCmd.Use != "myapp" {
		t.Error("Root command not configured correctly")
	}

	// Test greet command
	greetCmd := NewGreetCmd()
	if greetCmd.Use != "greet [name]" {
		t.Error("Greet command not configured correctly")
	}

	// Test version command
	versionCmd := NewVersionCmd()
	if versionCmd.Use != "version" {
		t.Error("Version command not configured correctly")
	}

	// Test command execution
	buf := new(bytes.Buffer)
	greetCmd.SetOut(buf)
	greetCmd.SetArgs([]string{"Alice"})
	if err := greetCmd.Execute(); err != nil {
		t.Errorf("Greet command execution failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hello, Alice!") {
		t.Errorf("Unexpected output: %s", output)
	}

	t.Log("✓ cobra CLI works!")
}
