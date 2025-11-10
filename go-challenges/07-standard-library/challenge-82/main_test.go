package main

import "testing"

func TestStrconv(t *testing.T) {
	num, err := StringToInt("42")
	if err != nil || num != 42 {
		t.Error("StringToInt failed")
	}
	
	if IntToString(42) != "42" {
		t.Error("IntToString failed")
	}
	
	f, err := ParseFloat("3.14")
	if err != nil || f < 3.13 || f > 3.15 {
		t.Error("ParseFloat failed")
	}
	
	if FormatFloat(3.14159) != "3.14" {
		t.Error("FormatFloat failed")
	}
	
	b, err := ParseBool("true")
	if err != nil || !b {
		t.Error("ParseBool failed")
	}
	
	t.Log("✓ strconv package works!")
}
