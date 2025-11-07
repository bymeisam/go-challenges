package main

import "testing"

func TestPolymorphism(t *testing.T) {
	file := &FileLogger{Filename: "app.log"}
	console := &ConsoleLogger{}

	LogMessage(file, "test1")
	LogMessage(console, "test2")

	if len(file.Messages) != 1 || file.Messages[0] != "test1" {
		t.Error("FileLogger failed")
	}
	if len(console.Messages) != 1 || console.Messages[0] != "test2" {
		t.Error("ConsoleLogger failed")
	}
	t.Log("✓ Polymorphism works!")
}
