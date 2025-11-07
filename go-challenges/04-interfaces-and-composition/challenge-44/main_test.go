package main

import "testing"

func TestInterfaceSegregation(t *testing.T) {
	robot := Robot{Name: "R2D2"}
	human := Human{Name: "Alice"}

	// Both can work
	DoWork(robot)
	DoWork(human)

	// Only human can eat
	var _ Eater = human
	// var _ Eater = robot // This would fail!

	t.Log("✓ Interface segregation works!")
}
