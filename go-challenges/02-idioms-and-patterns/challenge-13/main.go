package main

import "errors"

func SafeDivide(a, b float64) (float64, error) {
	// TODO: Return error if b == 0
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func ChainOperations(x int) error {
	// TODO: Perform operations, return first error encountered
	if x < 0 {
		return errors.New("negative input")
	}
	if x > 100 {
		return errors.New("input too large")
	}
	return nil
}

func MultiError(inputs []string) []error {
	// TODO: Process each input, collect errors
	var errs []error
	for _, input := range inputs {
		if input == "" {
			errs = append(errs, errors.New("empty string"))
		}
	}
	return errs
}

func main() {}
