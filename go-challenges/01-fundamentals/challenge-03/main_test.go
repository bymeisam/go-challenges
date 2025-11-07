package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestCreateMap(t *testing.T) {
	m := CreateMap()
	expected := map[string]int{"one": 1, "two": 2, "three": 3}

	if !reflect.DeepEqual(m, expected) {
		t.Errorf("CreateMap() = %v; want %v", m, expected)
	}
	t.Log("✓ Map created correctly!")
}

func TestAddToMap(t *testing.T) {
	m := make(map[string]int)
	AddToMap(m, "test", 42)

	if m["test"] != 42 {
		t.Errorf("After AddToMap, m[\"test\"] = %d; want 42", m["test"])
	}
	t.Log("✓ Add to map works!")
}

func TestGetFromMap(t *testing.T) {
	m := map[string]int{"exists": 100}

	val, ok := GetFromMap(m, "exists")
	if !ok || val != 100 {
		t.Errorf("GetFromMap(m, \"exists\") = (%d, %v); want (100, true)", val, ok)
	}

	val, ok = GetFromMap(m, "not_exists")
	if ok {
		t.Errorf("GetFromMap(m, \"not_exists\") should return ok=false")
	}
	t.Log("✓ Get from map works!")
}

func TestDeleteFromMap(t *testing.T) {
	m := map[string]int{"delete_me": 123}
	DeleteFromMap(m, "delete_me")

	if _, exists := m["delete_me"]; exists {
		t.Errorf("Key \"delete_me\" should be deleted from map")
	}
	t.Log("✓ Delete from map works!")
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := MapKeys(m)

	sort.Strings(keys)  // Sort for consistent comparison
	expected := []string{"a", "b", "c"}

	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("MapKeys(m) = %v; want %v", keys, expected)
	}
	t.Log("✓ Get map keys works!")
}
