package main

import (
	"errors"
	"fmt"
)

func AssertString(v interface{}) string {
	return v.(string) // May panic!
}

func SafeAssert(v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", errors.New("not a string")
	}
	return s, nil
}

type Closer interface {
	Close() error
}

func CheckCapability(v interface{}) bool {
	_, ok := v.(Closer)
	return ok
}

type MyCloser struct{}

func (m MyCloser) Close() error {
	return nil
}

func main() {}
