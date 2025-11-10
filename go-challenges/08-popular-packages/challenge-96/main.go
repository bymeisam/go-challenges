package main

import (
	"errors"
)

type Calculator interface {
	Add(a, b int) int
	Divide(a, b int) (int, error)
}

type SimpleCalculator struct{}

func (c *SimpleCalculator) Add(a, b int) int {
	return a + b
}

func (c *SimpleCalculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

type MathService struct {
	calc Calculator
}

func NewMathService(calc Calculator) *MathService {
	return &MathService{calc: calc}
}

func (s *MathService) Calculate(operation string, a, b int) (int, error) {
	switch operation {
	case "add":
		return s.calc.Add(a, b), nil
	case "divide":
		return s.calc.Divide(a, b)
	default:
		return 0, errors.New("unknown operation")
	}
}

func main() {}
