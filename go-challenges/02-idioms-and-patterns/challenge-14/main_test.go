package main

import "testing"

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s.Host != "localhost" || s.Port != 8080 {
		t.Errorf("Default server config incorrect: %+v", s)
	}
	t.Log("✓ NewServer with defaults works!")
}

func TestServerOptions(t *testing.T) {
	s := NewServer(
		WithHost("example.com"),
		WithPort(3000),
	)
	if s.Host != "example.com" || s.Port != 3000 {
		t.Errorf("Server options not applied: %+v", s)
	}
	t.Log("✓ Functional options work!")
}
