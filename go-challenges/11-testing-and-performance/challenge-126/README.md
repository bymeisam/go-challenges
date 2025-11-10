# Challenge 126: Table-Driven Tests

**Difficulty:** ⭐⭐ Medium
**Topic:** Testing & Performance
**Estimated Time:** 25-30 minutes

## 🎯 Learning Goals

- Understand the table-driven testing pattern in Go
- Learn how to structure test cases in a table format
- Practice writing comprehensive test suites with minimal code duplication
- Master testing multiple scenarios efficiently

## 📝 Description

Table-driven tests are a Go idiom for testing multiple scenarios with the same logic. Instead of writing separate test functions for each case, you define a slice of test cases (a "table") and loop through them. This approach:

1. **Reduces code duplication**: Write test logic once, apply to many cases
2. **Improves readability**: Test cases are clearly defined in a structured format
3. **Easy to extend**: Add new test cases by adding rows to the table
4. **Standard in Go**: This is the idiomatic way to test in Go

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. `Add(a, b int) int`

Add two integers and return the result.

### 2. `IsPalindrome(s string) bool`

Check if a string is a palindrome (reads the same forwards and backwards).
Ignore case and spaces.

### 3. `ValidateEmail(email string) bool`

Validate an email address:
- Must contain exactly one '@' symbol
- Must have at least one character before and after '@'
- Must have at least one '.' after '@'

### 4. `Fibonacci(n int) int`

Return the nth Fibonacci number (0-indexed).
- F(0) = 0
- F(1) = 1
- F(n) = F(n-1) + F(n-2)

## 🧪 Testing

The test file demonstrates table-driven testing patterns:

```bash
cd go-challenges/11-testing-and-performance/challenge-126
go test -v
```

All tests must pass! ✅

## 💡 Table-Driven Testing Pattern

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -1, -1, -2},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        result := Add(tt.a, tt.b)
        if result != tt.expected {
            t.Errorf("%s: Add(%d, %d) = %d; want %d",
                tt.name, tt.a, tt.b, result, tt.expected)
        }
    }
}
```

## 📚 Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go by Example: Testing](https://gobyexample.com/testing)

## ✨ Benefits of Table-Driven Tests

1. **DRY Principle**: Don't Repeat Yourself
2. **Easy to add test cases**: Just add another row
3. **Clear separation**: Test data vs test logic
4. **Better coverage**: Encourages testing edge cases

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
