package main

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	// Test UUID generation
	id1, err := GenerateUUIDv4()
	if err != nil {
		t.Fatalf("Failed to generate UUID: %v", err)
	}

	if !IsValidUUID(id1) {
		t.Error("Generated UUID is not valid")
	}

	// Test UUID parsing
	parsed, err := ParseUUID(id1)
	if err != nil {
		t.Errorf("Failed to parse UUID: %v", err)
	}

	if parsed.String() != id1 {
		t.Error("Parsed UUID doesn't match original")
	}

	// Test invalid UUID
	if IsValidUUID("invalid-uuid") {
		t.Error("Invalid UUID should not be valid")
	}

	// Test UUID version
	version, err := GetUUIDVersion(id1)
	if err != nil {
		t.Errorf("Failed to get UUID version: %v", err)
	}

	if version != 4 {
		t.Errorf("Expected UUID version 4, got %d", version)
	}

	// Test name-based UUID
	namespace := uuid.NameSpaceDNS
	name := "example.com"
	namedUUID := GenerateUUIDFromName(namespace, name)
	
	if !IsValidUUID(namedUUID) {
		t.Error("Named UUID is not valid")
	}

	// Name-based UUIDs should be deterministic
	namedUUID2 := GenerateUUIDFromName(namespace, name)
	if namedUUID != namedUUID2 {
		t.Error("Named UUIDs should be the same for same input")
	}

	// Test UUID to bytes and back
	bytes, err := UUIDToBytes(id1)
	if err != nil {
		t.Errorf("Failed to convert UUID to bytes: %v", err)
	}

	if len(bytes) != 16 {
		t.Errorf("Expected 16 bytes, got %d", len(bytes))
	}

	reconstructed, err := UUIDFromBytes(bytes)
	if err != nil {
		t.Errorf("Failed to reconstruct UUID from bytes: %v", err)
	}

	if reconstructed != id1 {
		t.Error("Reconstructed UUID doesn't match original")
	}

	t.Log("✓ uuid package works!")
}
