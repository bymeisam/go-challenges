package main

import "testing"

func TestParseInt(t *testing.T) {
	val, ok := ParseInt("42")
	if !ok || val != 42 {
		t.Errorf("ParseInt(\"42\") = (%d, %v); want (42, true)", val, ok)
	}
	_, ok = ParseInt("abc")
	if ok {
		t.Error("ParseInt(\"abc\") should return ok=false")
	}
	t.Log("✓ ParseInt works!")
}

func TestFindUser(t *testing.T) {
	user, err := FindUser(1)
	if err != nil || user == nil || user.ID != 1 {
		t.Errorf("FindUser(1) failed")
	}
	_, err = FindUser(999)
	if err == nil {
		t.Error("FindUser(999) should return error")
	}
	t.Log("✓ FindUser works!")
}

func TestSplit(t *testing.T) {
	first, last := Split("John Doe")
	if first != "John" || last != "Doe" {
		t.Errorf("Split(\"John Doe\") = (%q, %q); want (\"John\", \"Doe\")", first, last)
	}
	t.Log("✓ Split works!")
}
