package main

import (
	"reflect"
	"testing"
)

func TestDivide(t *testing.T) {
	result, err := Divide(10, 2)
	if err != nil || result != 5 {
		t.Errorf("Divide(10, 2) failed")
	}

	_, err = Divide(10, 0)
	if err == nil {
		t.Errorf("Divide by zero should return error")
	}
	t.Log("✓ Divide works!")
}

func TestMakeMultiplier(t *testing.T) {
	double := MakeMultiplier(2)
	if double(5) != 10 {
		t.Errorf("Multiplier failed")
	}
	t.Log("✓ MakeMultiplier works!")
}

func TestApplyOperation(t *testing.T) {
	nums := []int{1, 2, 3}
	result := ApplyOperation(nums, func(n int) int { return n * 2 })
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ApplyOperation failed")
	}
	t.Log("✓ ApplyOperation works!")
}
