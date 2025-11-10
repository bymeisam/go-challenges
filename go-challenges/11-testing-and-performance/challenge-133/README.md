# Challenge 133: Integration Tests

**Difficulty:** ⭐⭐⭐ Hard
**Topic:** Testing & Performance
**Estimated Time:** 40-45 minutes

## 🎯 Learning Goals

- Understand integration testing vs unit testing
- Learn to test with external services (databases, APIs)
- Master using test containers and mocks for integration tests
- Practice writing end-to-end test scenarios

## 📝 Description

Integration tests verify that different parts of your system work together correctly. Unlike unit tests that test individual functions, integration tests:

1. **Test multiple components**: Database + service + API
2. **Use real or simulated external services**: Databases, message queues, APIs
3. **Verify end-to-end scenarios**: Complete user workflows
4. **Run slower**: But catch integration issues

Common integration test patterns:
- **Test databases**: In-memory or containerized databases
- **HTTP servers**: Test with `httptest` package
- **External APIs**: Mock with `httptest.Server`
- **Skip with flags**: Use `-short` flag to skip slow tests

## 🔨 Your Task

Implement the following in `main.go`:

### 1. `UserRepository` interface and implementation

```go
type UserRepository interface {
    Create(user User) error
    GetByID(id int) (User, error)
    GetByEmail(email string) (User, error)
    Delete(id int) error
}
```

Implement `InMemoryUserRepository` for testing.

### 2. `UserService` struct

Business logic layer:
- `NewUserService(repo UserRepository) *UserService`
- `RegisterUser(name, email, password string) (User, error)` - validate and create user
- `Authenticate(email, password string) (User, error)` - check credentials
- `GetUserProfile(id int) (User, error)` - get user info

### 3. `HTTPServer` struct

HTTP API layer:
- `NewHTTPServer(service *UserService) *HTTPServer`
- `HandleRegister(w http.ResponseWriter, r *http.Request)` - POST /register
- `HandleLogin(w http.ResponseWriter, r *http.Request)` - POST /login
- `HandleProfile(w http.ResponseWriter, r *http.Request)` - GET /profile/:id

### 4. `User` struct

```go
type User struct {
    ID       int
    Name     string
    Email    string
    Password string // In production, store hashed!
}
```

### 5. `ExternalAPIClient` struct

Client for external API:
- `FetchUserData(userID int) (map[string]interface{}, error)`

## 🧪 Testing

The test file demonstrates:
- Testing repository layer
- Testing service layer with mock repository
- Testing HTTP handlers with `httptest`
- End-to-end integration tests
- Using `-short` flag to skip slow tests

```bash
# Run all tests
go test -v

# Skip slow integration tests
go test -v -short

# Run only integration tests
go test -v -run Integration
```

## 💡 Integration Testing Patterns

### HTTP Testing with httptest
```go
func TestHTTPHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/path", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
    }
}
```

### Mock External API
```go
func TestWithMockAPI(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id": 1, "name": "test"}`))
    }))
    defer server.Close()

    // Use server.URL as API endpoint
}
```

### Skip Slow Tests
```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Slow integration test code
}
```

## 🎯 Integration Test Levels

1. **Component Integration**: Service + Repository
2. **API Integration**: HTTP handlers + Service + Repository
3. **External Integration**: Your API + External services
4. **End-to-End**: Complete user workflows

## 📚 Resources

- [Testing HTTP Handlers](https://pkg.go.dev/net/http/httptest)
- [Integration Testing in Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)
- [Test Containers](https://www.testcontainers.org/)
- [Testing Flag: -short](https://pkg.go.dev/testing#Short)

## ✨ Best Practices

1. **Isolate tests**: Each test should be independent
2. **Use test fixtures**: Set up clean state for each test
3. **Test realistic scenarios**: User workflows, error cases
4. **Clean up resources**: Databases, HTTP servers, files
5. **Make tests fast**: Use in-memory databases when possible
6. **Use build tags**: Separate unit and integration tests

## 🏗️ Build Tags (Optional)

Separate integration tests:
```go
//go:build integration
// +build integration

package main
```

Run with: `go test -tags=integration`

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
