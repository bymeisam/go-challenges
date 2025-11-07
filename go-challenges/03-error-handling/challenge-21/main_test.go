package main

import "testing"

func TestValidationError(t *testing.T) {
	err := ValidationError{Field: "email", Message: "invalid format"}
	if err.Error() != "email: invalid format" {
		t.Errorf("ValidationError.Error() incorrect")
	}
	t.Log("✓ Custom error type works!")
}

func TestValidateAge(t *testing.T) {
	if err := ValidateAge(25); err != nil {
		t.Error("ValidateAge(25) should not error")
	}
	if err := ValidateAge(200); err == nil {
		t.Error("ValidateAge(200) should error")
	}
	t.Log("✓ ValidateAge works!")
}
