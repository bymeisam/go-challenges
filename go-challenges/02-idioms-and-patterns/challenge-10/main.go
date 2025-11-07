package main

import "errors"

// Divide performs integer division with named returns
func Divide(a, b int) (result int, remainder int) {
	// TODO: Calculate quotient and remainder
	// Use named returns
	return 0, 0
}

// ReadConfig simulates reading config with defaults on error
func ReadConfig() (host string, port int, err error) {
	// TODO: Return defaults: host="localhost", port=8080, err=nil
	return "", 0, nil
}

// ProcessData calculates sum and count using naked return
func ProcessData(data []int) (sum, count int) {
	// TODO: Calculate sum and count, use naked return
	return
}

func main() {}
