package main

import "testing"

func TestDependencyInjection(t *testing.T) {
	// Use mock repository for testing
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	user, err := service.GetUser(1)
	if err != nil || user.Name != "Test User" {
		t.Error("Dependency injection failed")
	}

	t.Log("✓ Dependency injection works!")
}

func TestRealRepository(t *testing.T) {
	// Can swap in real repository
	realRepo := NewDBUserRepository()
	realRepo.Save(&User{ID: 2, Name: "Real User"})

	service := NewUserService(realRepo)
	user, _ := service.GetUser(2)

	if user.Name != "Real User" {
		t.Error("Real repository failed")
	}

	t.Log("✓ Real repository works!")
}
