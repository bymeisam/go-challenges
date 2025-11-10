package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test with assert
func TestCalculatorAssert(t *testing.T) {
	calc := &SimpleCalculator{}
	
	result := calc.Add(2, 3)
	assert.Equal(t, 5, result, "Addition should work")
	
	result, err := calc.Divide(10, 2)
	assert.NoError(t, err, "Division should not error")
	assert.Equal(t, 5, result, "Division should work")
	
	_, err = calc.Divide(10, 0)
	assert.Error(t, err, "Division by zero should error")
}

// Test with require (stops on first failure)
func TestCalculatorRequire(t *testing.T) {
	calc := &SimpleCalculator{}
	
	result, err := calc.Divide(10, 2)
	require.NoError(t, err, "Division should not error")
	require.Equal(t, 5, result, "Division should work")
}

// Mock calculator
type MockCalculator struct {
	mock.Mock
}

func (m *MockCalculator) Add(a, b int) int {
	args := m.Called(a, b)
	return args.Int(0)
}

func (m *MockCalculator) Divide(a, b int) (int, error) {
	args := m.Called(a, b)
	return args.Int(0), args.Error(1)
}

// Test with mock
func TestMathServiceWithMock(t *testing.T) {
	mockCalc := new(MockCalculator)
	service := NewMathService(mockCalc)
	
	mockCalc.On("Add", 2, 3).Return(5)
	
	result, err := service.Calculate("add", 2, 3)
	assert.NoError(t, err)
	assert.Equal(t, 5, result)
	
	mockCalc.AssertExpectations(t)
}

// Test suite
type CalculatorTestSuite struct {
	suite.Suite
	calc *SimpleCalculator
}

func (suite *CalculatorTestSuite) SetupTest() {
	suite.calc = &SimpleCalculator{}
}

func (suite *CalculatorTestSuite) TestAdd() {
	result := suite.calc.Add(1, 1)
	suite.Equal(2, result, "1 + 1 should equal 2")
}

func (suite *CalculatorTestSuite) TestDivide() {
	result, err := suite.calc.Divide(6, 2)
	suite.NoError(err)
	suite.Equal(3, result, "6 / 2 should equal 3")
}

func TestCalculatorSuite(t *testing.T) {
	suite.Run(t, new(CalculatorTestSuite))
}
