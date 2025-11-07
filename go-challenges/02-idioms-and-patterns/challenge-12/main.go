package main

import (
	"errors"
	"strconv"
	"strings"
)

type User struct {
	ID   int
	Name string
}

func ParseInt(s string) (int, bool) {
	// TODO: Try to parse string to int, return value and success bool
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return val, true
}

func FindUser(id int) (*User, error) {
	// TODO: If id == 1, return &User{ID: 1, Name: "Alice"}, nil
	// Otherwise return nil, errors.New("user not found")
	return nil, nil
}

func Split(s string) (first, last string) {
	// TODO: Split on space, return first and last name
	// e.g., "John Doe" -> "John", "Doe"
	parts := strings.Split(s, " ")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

func main() {}
