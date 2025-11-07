package main

import "testing"

func TestNewDatabase(t *testing.T) {
	db, err := NewDatabase("localhost:5432")
	if err != nil || db == nil {
		t.Error("NewDatabase should succeed with valid conn string")
	}
	_, err = NewDatabase("")
	if err == nil {
		t.Error("NewDatabase should error with empty conn string")
	}
	t.Log("✓ NewDatabase works!")
}

func TestMustNewDatabase(t *testing.T) {
	db := MustNewDatabase("localhost:5432")
	if db == nil {
		t.Error("MustNewDatabase should return database")
	}
	t.Log("✓ MustNewDatabase works!")
}
