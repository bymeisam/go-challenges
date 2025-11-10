# Challenge 128: Test Fixtures

**Difficulty:** ⭐⭐⭐ Hard
**Topic:** Testing & Performance
**Estimated Time:** 30-35 minutes

## 🎯 Learning Goals

- Understand test fixtures and setup/teardown patterns
- Learn to use `TestMain` for test suite setup
- Master creating temporary files and directories for tests
- Practice cleaning up resources after tests

## 📝 Description

Test fixtures are resources needed for tests to run, such as:
- Database connections
- Temporary files and directories
- Test data
- Configuration

Go provides several ways to manage fixtures:

1. **TestMain**: Suite-level setup/teardown (runs once for all tests)
2. **Setup/Teardown functions**: Per-test setup/teardown
3. **defer**: Cleanup within individual tests
4. **Temporary files**: Using `os.MkdirTemp` and `t.TempDir()`

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. `FileStorage` struct

A simple file-based storage system:
- `New(baseDir string) *FileStorage` - create new storage
- `Save(filename, content string) error` - save content to file
- `Load(filename string) (string, error)` - load content from file
- `Delete(filename string) error` - delete a file
- `List() ([]string, error)` - list all files

### 2. `Database` struct (mock)

A mock database for testing:
- `Connect(dsn string) (*Database, error)` - connect to database
- `Close() error` - close connection
- `InsertUser(name, email string) (int, error)` - insert user, return ID
- `GetUser(id int) (User, error)` - get user by ID
- `DeleteUser(id int) error` - delete user

### 3. `User` struct

```go
type User struct {
    ID    int
    Name  string
    Email string
}
```

## 🧪 Testing

The test file demonstrates:
- Using `TestMain` for suite setup/teardown
- Using `t.TempDir()` for temporary directories
- Setup/teardown helper functions
- Cleanup with `defer` and `t.Cleanup()`

```bash
cd go-challenges/11-testing-and-performance/challenge-128
go test -v
```

All tests must pass! ✅

## 💡 Test Fixture Patterns

### TestMain Pattern
```go
func TestMain(m *testing.M) {
    // Setup
    setUp()

    // Run tests
    code := m.Run()

    // Teardown
    tearDown()

    os.Exit(code)
}
```

### Per-Test Setup with t.Cleanup()
```go
func TestExample(t *testing.T) {
    // Setup
    resource := setupResource()

    // Cleanup (runs after test)
    t.Cleanup(func() {
        cleanupResource(resource)
    })

    // Test code
}
```

### Temporary Directories
```go
func TestExample(t *testing.T) {
    // Creates temp dir, auto-cleaned after test
    dir := t.TempDir()

    // Use dir for test
}
```

## 🎯 Best Practices

1. **Always clean up**: Use `defer`, `t.Cleanup()`, or `TestMain`
2. **Isolate tests**: Each test should have its own fixtures
3. **Use temp dirs**: `t.TempDir()` for file operations
4. **Parallel safe**: Make fixtures safe for parallel tests
5. **Fail fast**: If setup fails, skip or fail the test immediately

## 📚 Resources

- [Testing Package: TestMain](https://pkg.go.dev/testing#hdr-Main)
- [Testing Package: Cleanup](https://pkg.go.dev/testing#T.Cleanup)
- [Testing Package: TempDir](https://pkg.go.dev/testing#T.TempDir)

## ✨ Why Test Fixtures Matter

- **Consistency**: Tests start from known state
- **Isolation**: Tests don't affect each other
- **Cleanup**: No leftover test artifacts
- **Reusability**: Share setup code across tests

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
