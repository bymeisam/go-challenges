package main

import "testing"

func TestNewPerson(t *testing.T) {
	p := NewPerson("Alice", 25)
	if p.Name != "Alice" || p.Age != 25 {
		t.Errorf("NewPerson failed: got %+v", p)
	}
	t.Log("✓ NewPerson works!")
}

func TestUpdateAge(t *testing.T) {
	p := Person{Name: "Bob", Age: 30}
	UpdateAge(&p, 31)
	if p.Age != 31 {
		t.Errorf("UpdateAge failed: age is %d, want 31", p.Age)
	}
	t.Log("✓ UpdateAge works!")
}

func TestGetInfo(t *testing.T) {
	p := Person{Name: "Charlie", Age: 40}
	info := GetInfo(p)
	expected := "Name: Charlie, Age: 40"
	if info != expected {
		t.Errorf("GetInfo = %q; want %q", info, expected)
	}
	t.Log("✓ GetInfo works!")
}
