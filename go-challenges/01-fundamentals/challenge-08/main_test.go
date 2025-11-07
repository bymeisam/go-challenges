package main

import (
	"strings"
	"testing"
)

func TestUserToJSON(t *testing.T) {
	user := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	jsonStr, err := UserToJSON(user)
	if err != nil {
		t.Fatalf("UserToJSON failed: %v", err)
	}
	if !strings.Contains(jsonStr, "Alice") {
		t.Errorf("JSON should contain 'Alice': %s", jsonStr)
	}
	t.Log("✓ UserToJSON works!")
}

func TestJSONToUser(t *testing.T) {
	jsonStr := `{"name":"Bob","email":"bob@example.com","age":25}`
	user, err := JSONToUser(jsonStr)
	if err != nil {
		t.Fatalf("JSONToUser failed: %v", err)
	}
	if user.Name != "Bob" || user.Age != 25 {
		t.Errorf("User not parsed correctly: %+v", user)
	}
	t.Log("✓ JSONToUser works!")
}

func TestRoundTrip(t *testing.T) {
	original := User{Name: "Charlie", Email: "charlie@example.com", Age: 35}
	jsonStr, _ := UserToJSON(original)
	decoded, _ := JSONToUser(jsonStr)

	if decoded.Name != original.Name || decoded.Age != original.Age {
		t.Errorf("Round trip failed")
	}
	t.Log("✓ Round trip works!")
}
