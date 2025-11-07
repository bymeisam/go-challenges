package main

import "testing"

func TestMocking(t *testing.T) {
	mock := &MockEmailSender{}
	service := NewNotificationService(mock)

	service.NotifyUser("test@example.com", "Hello")

	if len(mock.Calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(mock.Calls))
	}

	call := mock.Calls[0]
	if call.To != "test@example.com" || call.Body != "Hello" {
		t.Error("Mock didn't record call correctly")
	}

	t.Log("✓ Mocking works!")
}
