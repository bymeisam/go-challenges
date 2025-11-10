package main

import (
	"encoding/json"
	"testing"
)

func TestJSON(t *testing.T) {
	user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	
	data, err := MarshalUser(user)
	if err != nil {
		t.Fatal("MarshalUser failed")
	}
	
	decoded, err := UnmarshalUser(data)
	if err != nil || decoded.Name != "Alice" {
		t.Error("UnmarshalUser failed")
	}
	
	users := []User{user, {ID: 2, Name: "Bob"}}
	data, err = MarshalSlice(users)
	if err != nil {
		t.Error("MarshalSlice failed")
	}
	
	var result []User
	json.Unmarshal(data, &result)
	if len(result) != 2 {
		t.Error("Should marshal slice")
	}
	
	t.Log("✓ encoding/json works!")
}
