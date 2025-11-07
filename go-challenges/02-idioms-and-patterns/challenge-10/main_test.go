package main

import "testing"

func TestDivide(t *testing.T) {
	result, remainder := Divide(10, 3)
	if result != 3 || remainder != 1 {
		t.Errorf("Divide(10, 3) = (%d, %d); want (3, 1)", result, remainder)
	}
	t.Log("✓ Divide works!")
}

func TestReadConfig(t *testing.T) {
	host, port, err := ReadConfig()
	if host != "localhost" || port != 8080 || err != nil {
		t.Errorf("ReadConfig() = (%q, %d, %v); want (\"localhost\", 8080, nil)", host, port, err)
	}
	t.Log("✓ ReadConfig works!")
}

func TestProcessData(t *testing.T) {
	sum, count := ProcessData([]int{1, 2, 3, 4, 5})
	if sum != 15 || count != 5 {
		t.Errorf("ProcessData([1,2,3,4,5]) = (%d, %d); want (15, 5)", sum, count)
	}
	t.Log("✓ ProcessData works!")
}
