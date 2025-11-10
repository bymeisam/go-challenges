package main

import "testing"

func TestObserver(t *testing.T) {
	subject := &Subject{}
	email := &EmailNotifier{}
	sms := &SMSNotifier{}
	
	subject.Attach(email)
	subject.Attach(sms)
	
	subject.Notify("Hello observers")
	
	if len(email.Messages) != 1 || email.Messages[0] != "Email: Hello observers" {
		t.Error("Email observer failed")
	}
	
	if len(sms.Messages) != 1 || sms.Messages[0] != "SMS: Hello observers" {
		t.Error("SMS observer failed")
	}
	
	t.Log("✓ Observer pattern works!")
}
