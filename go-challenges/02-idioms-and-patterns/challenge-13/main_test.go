package main

import "testing"

func TestSafeDivide(t *testing.T) {
	result, err := SafeDivide(10, 2)
	if err != nil || result != 5 {
		t.Errorf("SafeDivide(10, 2) failed")
	}
	_, err = SafeDivide(10, 0)
	if err == nil {
		t.Error("SafeDivide by zero should return error")
	}
	t.Log("✓ SafeDivide works!")
}

func TestChainOperations(t *testing.T) {
	if err := ChainOperations(50); err != nil {
		t.Errorf("ChainOperations(50) should not error: %v", err)
	}
	if err := ChainOperations(-1); err == nil {
		t.Error("ChainOperations(-1) should error")
	}
	t.Log("✓ ChainOperations works!")
}

func TestMultiError(t *testing.T) {
	errs := MultiError([]string{"ok", "", "good", ""})
	if len(errs) != 2 {
		t.Errorf("MultiError should return 2 errors, got %d", len(errs))
	}
	t.Log("✓ MultiError works!")
}
