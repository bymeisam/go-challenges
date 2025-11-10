package main

import "testing"

func TestStrategy(t *testing.T) {
	cart := &ShoppingCart{}
	
	cart.SetPaymentStrategy(CreditCard{Number: "1234"})
	if cart.Checkout(100) != "Paid with credit card" {
		t.Error("Credit card strategy failed")
	}
	
	cart.SetPaymentStrategy(PayPal{Email: "test@example.com"})
	if cart.Checkout(100) != "Paid with PayPal" {
		t.Error("PayPal strategy failed")
	}
	
	t.Log("✓ Strategy pattern works!")
}
