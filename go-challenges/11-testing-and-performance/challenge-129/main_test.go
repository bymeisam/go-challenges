package main

import (
	"errors"
	"strings"
	"testing"
)

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	SendEmailFunc func(to, subject, body string) error
	Calls         []EmailCall
}

type EmailCall struct {
	To      string
	Subject string
	Body    string
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	m.Calls = append(m.Calls, EmailCall{To: to, Subject: subject, Body: body})
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, body)
	}
	return nil
}

// MockPaymentGateway is a mock implementation of PaymentGateway
type MockPaymentGateway struct {
	ProcessPaymentFunc func(amount float64, cardNumber string) (string, error)
	RefundPaymentFunc  func(transactionID string) error
	PaymentCalls       []PaymentCall
	RefundCalls        []RefundCall
}

type PaymentCall struct {
	Amount     float64
	CardNumber string
}

type RefundCall struct {
	TransactionID string
}

func (m *MockPaymentGateway) ProcessPayment(amount float64, cardNumber string) (string, error) {
	m.PaymentCalls = append(m.PaymentCalls, PaymentCall{Amount: amount, CardNumber: cardNumber})
	if m.ProcessPaymentFunc != nil {
		return m.ProcessPaymentFunc(amount, cardNumber)
	}
	return "txn_mock_123", nil
}

func (m *MockPaymentGateway) RefundPayment(transactionID string) error {
	m.RefundCalls = append(m.RefundCalls, RefundCall{TransactionID: transactionID})
	if m.RefundPaymentFunc != nil {
		return m.RefundPaymentFunc(transactionID)
	}
	return nil
}

func TestUserService_RegisterUser(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		mockEmail := &MockEmailService{}
		userService := NewUserService(mockEmail)

		err := userService.RegisterUser("Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("RegisterUser failed: %v", err)
		}

		// Verify email was sent
		if len(mockEmail.Calls) != 1 {
			t.Fatalf("expected 1 email call, got %d", len(mockEmail.Calls))
		}

		call := mockEmail.Calls[0]
		if call.To != "alice@example.com" {
			t.Errorf("email To = %q; want %q", call.To, "alice@example.com")
		}
		if call.Subject != "Welcome to our platform!" {
			t.Errorf("email Subject = %q; want %q", call.Subject, "Welcome to our platform!")
		}
		if !strings.Contains(call.Body, "Alice") {
			t.Errorf("email Body should contain user name, got: %q", call.Body)
		}
	})

	t.Run("email service failure", func(t *testing.T) {
		mockEmail := &MockEmailService{
			SendEmailFunc: func(to, subject, body string) error {
				return errors.New("email service unavailable")
			},
		}
		userService := NewUserService(mockEmail)

		err := userService.RegisterUser("Bob", "bob@example.com")
		if err == nil {
			t.Fatal("expected error when email service fails")
		}

		if !strings.Contains(err.Error(), "failed to send welcome email") {
			t.Errorf("error should mention email failure, got: %v", err)
		}
	})

	t.Run("empty name or email", func(t *testing.T) {
		mockEmail := &MockEmailService{}
		userService := NewUserService(mockEmail)

		tests := []struct {
			name  string
			email string
		}{
			{"", "test@example.com"},
			{"Test", ""},
			{"", ""},
		}

		for _, tt := range tests {
			err := userService.RegisterUser(tt.name, tt.email)
			if err == nil {
				t.Errorf("RegisterUser(%q, %q) should return error", tt.name, tt.email)
			}
		}
	})

	t.Log("✓ All UserService RegisterUser tests passed!")
}

func TestUserService_ResetPassword(t *testing.T) {
	t.Run("successful password reset", func(t *testing.T) {
		mockEmail := &MockEmailService{}
		userService := NewUserService(mockEmail)

		err := userService.ResetPassword("user@example.com", "newpassword123")
		if err != nil {
			t.Fatalf("ResetPassword failed: %v", err)
		}

		// Verify notification email was sent
		if len(mockEmail.Calls) != 1 {
			t.Fatalf("expected 1 email call, got %d", len(mockEmail.Calls))
		}

		call := mockEmail.Calls[0]
		if call.To != "user@example.com" {
			t.Errorf("email To = %q; want %q", call.To, "user@example.com")
		}
		if call.Subject != "Password Reset Successful" {
			t.Errorf("email Subject = %q; want %q", call.Subject, "Password Reset Successful")
		}
	})

	t.Run("password too short", func(t *testing.T) {
		mockEmail := &MockEmailService{}
		userService := NewUserService(mockEmail)

		err := userService.ResetPassword("user@example.com", "short")
		if err == nil {
			t.Fatal("expected error for short password")
		}

		// Should not send email for validation error
		if len(mockEmail.Calls) != 0 {
			t.Error("should not send email when validation fails")
		}
	})

	t.Log("✓ All UserService ResetPassword tests passed!")
}

func TestOrderService_PlaceOrder(t *testing.T) {
	t.Run("successful order", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{
			ProcessPaymentFunc: func(amount float64, cardNumber string) (string, error) {
				return "txn_12345", nil
			},
		}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		transactionID, err := orderService.PlaceOrder("customer@example.com", 99.99, "4111111111111111")
		if err != nil {
			t.Fatalf("PlaceOrder failed: %v", err)
		}

		if transactionID != "txn_12345" {
			t.Errorf("transactionID = %q; want %q", transactionID, "txn_12345")
		}

		// Verify payment was processed
		if len(mockPayment.PaymentCalls) != 1 {
			t.Fatalf("expected 1 payment call, got %d", len(mockPayment.PaymentCalls))
		}

		paymentCall := mockPayment.PaymentCalls[0]
		if paymentCall.Amount != 99.99 {
			t.Errorf("payment amount = %.2f; want %.2f", paymentCall.Amount, 99.99)
		}

		// Verify confirmation email was sent
		if len(mockEmail.Calls) != 1 {
			t.Fatalf("expected 1 email call, got %d", len(mockEmail.Calls))
		}

		emailCall := mockEmail.Calls[0]
		if emailCall.To != "customer@example.com" {
			t.Errorf("email To = %q; want %q", emailCall.To, "customer@example.com")
		}
		if !strings.Contains(emailCall.Body, "txn_12345") {
			t.Error("confirmation email should contain transaction ID")
		}
	})

	t.Run("payment failure", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{
			ProcessPaymentFunc: func(amount float64, cardNumber string) (string, error) {
				return "", errors.New("insufficient funds")
			},
		}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		_, err := orderService.PlaceOrder("customer@example.com", 99.99, "4111111111111111")
		if err == nil {
			t.Fatal("expected error when payment fails")
		}

		// Should not send email when payment fails
		if len(mockEmail.Calls) != 0 {
			t.Error("should not send email when payment fails")
		}
	})

	t.Run("email failure after successful payment", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{}
		mockEmail := &MockEmailService{
			SendEmailFunc: func(to, subject, body string) error {
				return errors.New("email service down")
			},
		}
		orderService := NewOrderService(mockPayment, mockEmail)

		transactionID, err := orderService.PlaceOrder("customer@example.com", 99.99, "4111111111111111")

		// Should still return transaction ID even if email fails
		if transactionID == "" {
			t.Error("should return transaction ID even if email fails")
		}

		// Should return error mentioning email failure
		if err == nil {
			t.Fatal("expected error when email fails")
		}
		if !strings.Contains(err.Error(), "confirmation email failed") {
			t.Errorf("error should mention email failure, got: %v", err)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		tests := []struct {
			name       string
			email      string
			amount     float64
			cardNumber string
		}{
			{"empty email", "", 99.99, "4111111111111111"},
			{"zero amount", "test@example.com", 0, "4111111111111111"},
			{"negative amount", "test@example.com", -10, "4111111111111111"},
			{"empty card", "test@example.com", 99.99, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := orderService.PlaceOrder(tt.email, tt.amount, tt.cardNumber)
				if err == nil {
					t.Error("expected error for invalid input")
				}
			})
		}
	})

	t.Log("✓ All OrderService PlaceOrder tests passed!")
}

func TestOrderService_CancelOrder(t *testing.T) {
	t.Run("successful cancellation", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		err := orderService.CancelOrder("order_123", "txn_12345")
		if err != nil {
			t.Fatalf("CancelOrder failed: %v", err)
		}

		// Verify refund was processed
		if len(mockPayment.RefundCalls) != 1 {
			t.Fatalf("expected 1 refund call, got %d", len(mockPayment.RefundCalls))
		}

		refundCall := mockPayment.RefundCalls[0]
		if refundCall.TransactionID != "txn_12345" {
			t.Errorf("refund transaction ID = %q; want %q", refundCall.TransactionID, "txn_12345")
		}
	})

	t.Run("refund failure", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{
			RefundPaymentFunc: func(transactionID string) error {
				return errors.New("refund not allowed")
			},
		}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		err := orderService.CancelOrder("order_123", "txn_12345")
		if err == nil {
			t.Fatal("expected error when refund fails")
		}
	})

	t.Run("empty parameters", func(t *testing.T) {
		mockPayment := &MockPaymentGateway{}
		mockEmail := &MockEmailService{}
		orderService := NewOrderService(mockPayment, mockEmail)

		tests := []struct {
			orderID       string
			transactionID string
		}{
			{"", "txn_123"},
			{"order_123", ""},
			{"", ""},
		}

		for _, tt := range tests {
			err := orderService.CancelOrder(tt.orderID, tt.transactionID)
			if err == nil {
				t.Errorf("CancelOrder(%q, %q) should return error", tt.orderID, tt.transactionID)
			}
		}
	})

	t.Log("✓ All OrderService CancelOrder tests passed!")
}
