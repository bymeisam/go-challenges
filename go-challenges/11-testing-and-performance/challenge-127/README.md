# Challenge 127: Subtests

**Difficulty:** ⭐⭐ Medium
**Topic:** Testing & Performance
**Estimated Time:** 25-30 minutes

## 🎯 Learning Goals

- Master the `t.Run()` function for creating subtests
- Learn how to organize complex test suites
- Understand parallel test execution with `t.Parallel()`
- Practice running specific subtests using `-run` flag

## 📝 Description

Subtests allow you to group related tests and run them independently. They provide better organization, clearer output, and the ability to run specific test cases. Using `t.Run()`:

1. **Better organization**: Group related test cases logically
2. **Clearer output**: Each subtest has its own name in test results
3. **Selective execution**: Run specific subtests with `-run` flag
4. **Parallel execution**: Subtests can run in parallel with `t.Parallel()`

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. `Calculator` struct with methods

Create a calculator that supports:
- `Add(a, b int) int`
- `Subtract(a, b int) int`
- `Multiply(a, b int) int`
- `Divide(a, b float64) (float64, error)` - return error for division by zero

### 2. `UserValidator` struct

Validate user registration data:
- `ValidateUsername(username string) error` - must be 3-20 chars, alphanumeric
- `ValidatePassword(password string) error` - must be 8+ chars with at least one number
- `ValidateAge(age int) error` - must be 18 or older

### 3. `StringProcessor` struct

String processing utilities:
- `Reverse(s string) string` - reverse a string
- `WordCount(s string) int` - count words in a string
- `ToSnakeCase(s string) string` - convert to snake_case (e.g., "Hello World" -> "hello_world")

## 🧪 Testing

Run all tests:
```bash
go test -v
```

Run specific subtests:
```bash
# Run all Calculator tests
go test -v -run TestCalculator

# Run only Calculator/Addition subtests
go test -v -run TestCalculator/Addition

# Run tests in parallel
go test -v -parallel 4
```

All tests must pass! ✅

## 💡 Subtests Pattern

```go
func TestCalculator(t *testing.T) {
    calc := Calculator{}

    t.Run("Addition", func(t *testing.T) {
        result := calc.Add(2, 3)
        if result != 5 {
            t.Errorf("got %d, want 5", result)
        }
    })

    t.Run("Subtraction", func(t *testing.T) {
        result := calc.Subtract(5, 3)
        if result != 2 {
            t.Errorf("got %d, want 2", result)
        }
    })
}
```

## 🎯 Key Features of Subtests

1. **Named test cases**: Each subtest has a descriptive name
2. **Isolation**: Each subtest is independent
3. **Setup/Teardown**: Can have per-subtest setup
4. **Parallel execution**: Use `t.Parallel()` for concurrent tests
5. **Selective running**: `-run` flag to run specific tests

## 📚 Resources

- [Go Testing: Subtests](https://go.dev/blog/subtests)
- [Testing Package Documentation](https://pkg.go.dev/testing)
- [Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)

## ✨ Bonus Challenge (Optional)

Try running tests in parallel by adding `t.Parallel()` at the start of subtests. Compare execution time!

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
