package main

import (
	"testing"
)

func TestValidator(t *testing.T) {
	// Valid user
	validUser := &User{
		Email:    "user@example.com",
		Age:      25,
		Password: "password123",
		Website:  "https://example.com",
	}

	if err := ValidateUser(validUser); err != nil {
		t.Errorf("Valid user should not have validation errors: %v", err)
	}

	// Invalid email
	invalidUser := &User{
		Email:    "invalid-email",
		Age:      25,
		Password: "password123",
	}

	err := ValidateUser(invalidUser)
	if err == nil {
		t.Error("Invalid email should fail validation")
	}

	errors := GetValidationErrors(err)
	if _, exists := errors["Email"]; !exists {
		t.Error("Expected Email field to have validation error")
	}

	// Invalid age
	youngUser := &User{
		Email:    "user@example.com",
		Age:      16,
		Password: "password123",
	}

	err = ValidateUser(youngUser)
	if err == nil {
		t.Error("Age below 18 should fail validation")
	}

	// Invalid password
	weakPassword := &User{
		Email:    "user@example.com",
		Age:      25,
		Password: "123",
	}

	err = ValidateUser(weakPassword)
	if err == nil {
		t.Error("Short password should fail validation")
	}

	// Valid address
	validAddress := &Address{
		Street:  "123 Main St",
		City:    "New York",
		Country: "US",
		ZipCode: "10001",
	}

	if err := ValidateAddress(validAddress); err != nil {
		t.Errorf("Valid address should not have validation errors: %v", err)
	}

	// Invalid zip code
	invalidAddress := &Address{
		Street:  "123 Main St",
		City:    "New York",
		Country: "US",
		ZipCode: "ABC",
	}

	err = ValidateAddress(invalidAddress)
	if err == nil {
		t.Error("Invalid zip code should fail validation")
	}

	t.Log("✓ validator package works!")
}
