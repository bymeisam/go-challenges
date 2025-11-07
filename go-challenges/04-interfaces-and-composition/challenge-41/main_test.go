package main

import (
	"strings"
	"testing"
)

func TestAbstractFactory(t *testing.T) {
	db := NewDatabase("mysql", "localhost")
	if db == nil {
		t.Fatal("Factory should create database")
	}
	if !strings.Contains(db.Connect(), "MySQL") {
		t.Error("Should create MySQL database")
	}

	db2 := NewDatabase("postgres", "localhost")
	if !strings.Contains(db2.Connect(), "PostgreSQL") {
		t.Error("Should create PostgreSQL database")
	}

	t.Log("✓ Abstract Factory pattern works!")
}
