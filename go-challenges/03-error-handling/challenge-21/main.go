package main

import "fmt"

type ValidationError struct {
	Field string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateAge(age int) error {
	if age < 0 || age > 150 {
		return ValidationError{Field: "age", Message: "invalid age"}
	}
	return nil
}

func main() {}
