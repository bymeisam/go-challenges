package main

import (
	"errors"
	"fmt"
	"os"
)

func ReadFile(filename string) error {
	_, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}
	return nil
}

var ErrNotFound = errors.New("not found")

func FindItem(id int) error {
	if id == 0 {
		return fmt.Errorf("item %d: %w", id, ErrNotFound)
	}
	return nil
}

func main() {}
