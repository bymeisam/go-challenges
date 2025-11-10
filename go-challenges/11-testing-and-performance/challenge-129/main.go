package main

import (
	"errors"
	"fmt"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// PaymentGateway defines the interface for payment processing
type PaymentGateway interface {
	ProcessPayment(amount float64, cardNumber string) (string, error)
	RefundPayment(transactionID string) error
}

// UserService handles user-related operations
type UserService struct {
	emailService EmailService
}

// NewUserService creates a new UserService with email dependency
func NewUserService(emailService EmailService) *UserService {
	return &UserService{
		emailService: emailService,
	}
}

// RegisterUser registers a new user and sends a welcome email
func (us *UserService) RegisterUser(name, email string) error {
	if name == "" || email == "" {
		return errors.New("name and email are required")
	}

	// In a real implementation, we would save to database here

	// Send welcome email
	subject := "Welcome to our platform!"
	body := fmt.Sprintf("Hello %s, welcome to our platform!", name)

	if err := us.emailService.SendEmail(email, subject, body); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password and sends notification
func (us *UserService) ResetPassword(email, newPassword string) error {
	if email == "" || newPassword == "" {
		return errors.New("email and new password are required")
	}

	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// In a real implementation, we would update database here

	// Send password reset notification
	subject := "Password Reset Successful"
	body := "Your password has been successfully reset."

	if err := us.emailService.SendEmail(email, subject, body); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// OrderService handles order-related operations
type OrderService struct {
	paymentGateway PaymentGateway
	emailService   EmailService
}

// NewOrderService creates a new OrderService with payment and email dependencies
func NewOrderService(payment PaymentGateway, email EmailService) *OrderService {
	return &OrderService{
		paymentGateway: payment,
		emailService:   email,
	}
}

// PlaceOrder processes payment and sends confirmation email
func (os *OrderService) PlaceOrder(userEmail string, amount float64, cardNumber string) (string, error) {
	if userEmail == "" {
		return "", errors.New("user email is required")
	}

	if amount <= 0 {
		return "", errors.New("amount must be greater than zero")
	}

	if cardNumber == "" {
		return "", errors.New("card number is required")
	}

	// Process payment
	transactionID, err := os.paymentGateway.ProcessPayment(amount, cardNumber)
	if err != nil {
		return "", fmt.Errorf("payment failed: %w", err)
	}

	// Send order confirmation email
	subject := "Order Confirmation"
	body := fmt.Sprintf("Your order has been confirmed. Transaction ID: %s, Amount: $%.2f",
		transactionID, amount)

	if err := os.emailService.SendEmail(userEmail, subject, body); err != nil {
		// Payment succeeded but email failed - log this but don't fail the order
		// In production, you might want to retry or queue the email
		return transactionID, fmt.Errorf("order placed but confirmation email failed: %w", err)
	}

	return transactionID, nil
}

// CancelOrder refunds payment and sends cancellation email
func (os *OrderService) CancelOrder(orderID, transactionID string) error {
	if orderID == "" || transactionID == "" {
		return errors.New("order ID and transaction ID are required")
	}

	// Refund payment
	if err := os.paymentGateway.RefundPayment(transactionID); err != nil {
		return fmt.Errorf("refund failed: %w", err)
	}

	// In a real implementation, we would update order status in database here

	return nil
}

// RealEmailService is a real implementation (not used in tests)
type RealEmailService struct{}

func (res *RealEmailService) SendEmail(to, subject, body string) error {
	// In production, this would connect to an SMTP server
	fmt.Printf("Sending email to %s: %s\n", to, subject)
	return nil
}

// RealPaymentGateway is a real implementation (not used in tests)
type RealPaymentGateway struct{}

func (rpg *RealPaymentGateway) ProcessPayment(amount float64, cardNumber string) (string, error) {
	// In production, this would connect to a payment API
	fmt.Printf("Processing payment: $%.2f\n", amount)
	return "txn_123456", nil
}

func (rpg *RealPaymentGateway) RefundPayment(transactionID string) error {
	// In production, this would connect to a payment API
	fmt.Printf("Refunding payment: %s\n", transactionID)
	return nil
}

func main() {
	// Example usage with real implementations
	emailService := &RealEmailService{}
	paymentGateway := &RealPaymentGateway{}

	userService := NewUserService(emailService)
	orderService := NewOrderService(paymentGateway, emailService)

	// Register user
	userService.RegisterUser("John Doe", "john@example.com")

	// Place order
	orderService.PlaceOrder("john@example.com", 99.99, "4111111111111111")
}
