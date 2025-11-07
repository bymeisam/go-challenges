package main

import (
	"reflect"
	"testing"
)

func TestCreateArray(t *testing.T) {
	arr := CreateArray()
	expected := [5]int{1, 2, 3, 4, 5}

	if arr != expected {
		t.Errorf("CreateArray() = %v; want %v", arr, expected)
	}

	// Verify it's actually an array type
	if len(arr) != 5 {
		t.Errorf("Array length should be 5, got %d", len(arr))
	}

	t.Log("✓ Array created correctly!")
}

func TestCreateSlice(t *testing.T) {
	slice := CreateSlice()
	expected := []int{1, 2, 3, 4, 5}

	if !reflect.DeepEqual(slice, expected) {
		t.Errorf("CreateSlice() = %v; want %v", slice, expected)
	}

	if len(slice) != 5 {
		t.Errorf("Slice length should be 5, got %d", len(slice))
	}

	t.Log("✓ Slice created correctly!")
}

func TestAppendToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		value    int
		expected []int
	}{
		{"append to empty", []int{}, 1, []int{1}},
		{"append to single element", []int{1}, 2, []int{1, 2}},
		{"append to multiple elements", []int{1, 2, 3}, 4, []int{1, 2, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AppendToSlice(tt.input, tt.value)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("AppendToSlice(%v, %d) = %v; want %v",
					tt.input, tt.value, result, tt.expected)
			}
		})
	}

	t.Log("✓ Append works correctly!")
}

func TestSliceCapacity(t *testing.T) {
	tests := []struct {
		name            string
		slice           []int
		expectedLen     int
		expectedCapMin  int // capacity can vary
	}{
		{"empty slice", []int{}, 0, 0},
		{"small slice", []int{1, 2, 3}, 3, 3},
		{"slice from make", make([]int, 5, 10), 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, capacity := SliceCapacity(tt.slice)

			if length != tt.expectedLen {
				t.Errorf("SliceCapacity(%v) length = %d; want %d",
					tt.slice, length, tt.expectedLen)
			}

			if capacity < tt.expectedCapMin {
				t.Errorf("SliceCapacity(%v) capacity = %d; want at least %d",
					tt.slice, capacity, tt.expectedCapMin)
			}
		})
	}

	t.Log("✓ Length and capacity calculated correctly!")
}

func TestSlicePortions(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		firstPart []int
		secondPart []int
	}{
		{
			"standard case",
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2},
			[]int{3, 4, 5},
		},
		{
			"four elements",
			[]int{10, 20, 30, 40},
			[]int{10, 20},
			[]int{30, 40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, second := SlicePortions(tt.input)

			if !reflect.DeepEqual(first, tt.firstPart) {
				t.Errorf("SlicePortions(%v) first part = %v; want %v",
					tt.input, first, tt.firstPart)
			}

			if !reflect.DeepEqual(second, tt.secondPart) {
				t.Errorf("SlicePortions(%v) second part = %v; want %v",
					tt.input, second, tt.secondPart)
			}
		})
	}

	t.Log("✓ Slice portions work correctly!")
}

func BenchmarkAppendToSlice(b *testing.B) {
	slice := []int{1, 2, 3}
	for i := 0; i < b.N; i++ {
		AppendToSlice(slice, 4)
	}
}
