package main

import (
	"fmt"
	"testing"
)

func TestStringer(t *testing.T) {
	p := Product{Name: "Laptop", Price: 999.99}
	expected := "Laptop ($999.99)"
	if fmt.Sprint(p) != expected {
		t.Errorf("Stringer failed: got %s", p.String())
	}
	t.Log("✓ Stringer works!")
}

func TestError(t *testing.T) {
	err := OutOfStockError{Product: "Widget"}
	if err.Error() != "product Widget is out of stock" {
		t.Error("Error interface failed")
	}
	t.Log("✓ Error interface works!")
}
