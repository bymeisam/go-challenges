package main

import "testing"

func TestRepository(t *testing.T) {
	repo := NewInMemoryUserRepository()
	
	user := User{ID: 1, Name: "Alice"}
	repo.Save(user)
	
	found, _ := repo.FindByID(1)
	if found == nil || found.Name != "Alice" {
		t.Error("Repository save/find failed")
	}
	
	users, _ := repo.FindAll()
	if len(users) != 1 {
		t.Error("FindAll failed")
	}
	
	t.Log("✓ Repository pattern works!")
}
