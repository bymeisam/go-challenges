package main

import "errors"

var ErrInvalid = errors.New("invalid input")

func ProcessInput(s string) error {
	if s == "" {
		return ErrInvalid
	}
	return nil
}

func main() {}
