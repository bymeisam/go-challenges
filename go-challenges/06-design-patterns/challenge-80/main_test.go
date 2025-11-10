package main

import "testing"

func TestObjectPool(t *testing.T) {
	pool := NewPool(2)
	
	conn1 := pool.Acquire()
	if conn1 == nil {
		t.Fatal("Should acquire connection")
	}
	
	conn2 := pool.Acquire()
	if conn2 == nil {
		t.Fatal("Should acquire second connection")
	}
	
	conn3 := pool.Acquire()
	if conn3 != nil {
		t.Error("Pool should be empty")
	}
	
	pool.Release(conn1)
	conn4 := pool.Acquire()
	if conn4 == nil {
		t.Error("Should reuse released connection")
	}
	
	t.Log("✓ Object pool works!")
}
