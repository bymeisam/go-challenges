# Challenge 129: Mocking

**Difficulty:** ⭐⭐⭐ Hard
**Topic:** Testing & Performance
**Estimated Time:** 35-40 minutes

## 🎯 Learning Goals

- Understand mocking and why it's important for testing
- Learn to use interfaces for dependency injection
- Master creating mock implementations for testing
- Practice testing code with external dependencies

## 📝 Description

Mocking allows you to test code in isolation by replacing dependencies with controlled test doubles. In Go, mocking is typically done through:

1. **Interfaces**: Define contracts for dependencies
2. **Mock implementations**: Create test versions of interfaces
3. **Dependency injection**: Pass dependencies as parameters
4. **Table-driven mocks**: Different mock behaviors for different tests

Benefits of mocking:
- **Isolation**: Test units independently
- **Control**: Simulate different scenarios (success, failure, edge cases)
- **Speed**: No need for actual databases, APIs, etc.
- **Reliability**: Tests don't depend on external services

## 🔨 Your Task

Implement the following in `main.go`:

### 1. `EmailService` interface

```go
type EmailService interface {
    SendEmail(to, subject, body string) error
}
```

### 2. `PaymentGateway` interface

```go
type PaymentGateway interface {
    ProcessPayment(amount float64, cardNumber string) (string, error)
    RefundPayment(transactionID string) error
}
```

### 3. `UserService` struct

User service that depends on EmailService:
- `NewUserService(emailService EmailService) *UserService`
- `RegisterUser(name, email string) error` - registers user and sends welcome email
- `ResetPassword(email, newPassword string) error` - resets password and sends notification

### 4. `OrderService` struct

Order service that depends on PaymentGateway and EmailService:
- `NewOrderService(payment PaymentGateway, email EmailService) *OrderService`
- `PlaceOrder(userEmail string, amount float64, cardNumber string) (string, error)` - process payment and send confirmation
- `CancelOrder(orderID, transactionID string) error` - refund payment and send cancellation email

## 🧪 Testing

The test file demonstrates:
- Creating mock implementations
- Testing with different mock behaviors
- Verifying mock interactions
- Testing error scenarios

```bash
cd go-challenges/11-testing-and-performance/challenge-129
go test -v
```

All tests must pass! ✅

## 💡 Mocking Pattern

```go
// Define interface
type EmailService interface {
    SendEmail(to, subject, body string) error
}

// Mock implementation for testing
type MockEmailService struct {
    SendEmailFunc func(to, subject, body string) error
    Calls         []EmailCall
}

type EmailCall struct {
    To, Subject, Body string
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
    m.Calls = append(m.Calls, EmailCall{to, subject, body})
    if m.SendEmailFunc != nil {
        return m.SendEmailFunc(to, subject, body)
    }
    return nil
}
```

## 🎯 Mocking Best Practices

1. **Use interfaces**: Make code testable through interfaces
2. **Inject dependencies**: Pass dependencies as parameters
3. **Track calls**: Record mock interactions for verification
4. **Configurable behavior**: Allow tests to control mock responses
5. **Keep mocks simple**: Don't make mocks too complex

## 📚 Resources

- [Go Interfaces](https://go.dev/tour/methods/9)
- [Dependency Injection in Go](https://www.youtube.com/watch?v=EP_mDQ4NwfU)
- [Testing with Mocks](https://www.youtube.com/watch?v=8hQG7QlcLBk)
- [Testify Mock Library](https://github.com/stretchr/testify) (optional, not used here)

## ✨ Why Mocking Matters

Without mocking:
- Tests require real email servers, payment gateways
- Tests are slow and unreliable
- Hard to test error scenarios
- External dependencies must be available

With mocking:
- Fast, isolated unit tests
- Easy to test all scenarios
- No external dependencies
- Predictable, reliable tests

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
