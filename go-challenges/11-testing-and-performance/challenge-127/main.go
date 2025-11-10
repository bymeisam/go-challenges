package main

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

func (c Calculator) Add(a, b int) int {
	return a + b
}

func (c Calculator) Subtract(a, b int) int {
	return a - b
}

func (c Calculator) Multiply(a, b int) int {
	return a * b
}

func (c Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// UserValidator validates user registration data
type UserValidator struct{}

func (v UserValidator) ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be 3-20 characters")
	}

	alphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumeric.MatchString(username) {
		return errors.New("username must be alphanumeric")
	}

	return nil
}

func (v UserValidator) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	hasNumber := false
	for _, ch := range password {
		if unicode.IsDigit(ch) {
			hasNumber = true
			break
		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

func (v UserValidator) ValidateAge(age int) error {
	if age < 18 {
		return errors.New("must be 18 or older")
	}
	return nil
}

// StringProcessor provides string utility functions
type StringProcessor struct{}

func (sp StringProcessor) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (sp StringProcessor) WordCount(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	words := strings.Fields(s)
	return len(words)
}

func (sp StringProcessor) ToSnakeCase(s string) string {
	var result strings.Builder

	for i, ch := range s {
		if unicode.IsSpace(ch) {
			result.WriteRune('_')
		} else if unicode.IsUpper(ch) {
			if i > 0 && !unicode.IsSpace(rune(s[i-1])) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(ch))
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}

func main() {
	// Example usage
}
