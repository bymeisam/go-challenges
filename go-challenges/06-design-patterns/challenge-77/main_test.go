package main

import "testing"

func TestServiceLayer(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)
	
	user, err := service.RegisterUser("Bob")
	if err != nil || user.Name != "Bob" {
		t.Error("RegisterUser failed")
	}
	
	found, _ := service.GetUser(user.ID)
	if found == nil || found.Name != "Bob" {
		t.Error("GetUser failed")
	}
	
	t.Log("✓ Service layer works!")
}
