package main

import "testing"

func TestSortPeople(t *testing.T) {
	people := []Person{
		{"Bob", 30},
		{"Alice", 25},
		{"Charlie", 35},
	}
	SortPeople(people)
	if people[0].Name != "Alice" || people[0].Age != 25 {
		t.Error("Sort failed")
	}
	t.Log("✓ sort.Interface works!")
}
