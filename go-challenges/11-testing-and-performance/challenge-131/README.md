# Challenge 131: Race Detection

**Difficulty:** ⭐⭐⭐ Hard
**Topic:** Testing & Performance
**Estimated Time:** 30-35 minutes

## 🎯 Learning Goals

- Understand data races and why they're dangerous
- Learn to detect races with Go's race detector (`-race` flag)
- Master synchronization techniques to prevent races
- Practice writing race-free concurrent code

## 📝 Description

A data race occurs when two goroutines access the same variable concurrently, and at least one access is a write. Data races are bugs that can cause:
- Unpredictable behavior
- Crashes
- Data corruption
- Subtle bugs that are hard to reproduce

Go's race detector (`-race` flag) instruments your code to detect races at runtime.

## 🔨 Your Task

Implement the following in `main.go`:

### 1. `Counter` struct (unsafe version)

A counter that has race conditions:
- `NewCounter() *Counter`
- `Increment()` - increment counter (has race!)
- `Value() int` - get current value (has race!)

### 2. `SafeCounter` struct (thread-safe version)

A counter using mutex for safety:
- `NewSafeCounter() *SafeCounter`
- `Increment()` - thread-safe increment
- `Value() int` - thread-safe read

### 3. `Cache` struct (unsafe version)

A cache with race conditions:
- `NewCache() *Cache`
- `Set(key string, value interface{})` - has race!
- `Get(key string) (interface{}, bool)` - has race!

### 4. `SafeCache` struct (thread-safe version)

A cache using RWMutex:
- `NewSafeCache() *SafeCache`
- `Set(key string, value interface{})` - thread-safe write
- `Get(key string) (interface{}, bool)` - thread-safe read

### 5. `URLFetcher` (demonstrates proper synchronization)

- `FetchURLs(urls []string) map[string]string` - fetch URLs concurrently (safe)

## 🧪 Testing

Run tests normally:
```bash
go test -v
```

Run with race detector:
```bash
# This will detect race conditions!
go test -race -v

# Run specific test with race detector
go test -race -v -run TestCounter
```

The unsafe versions should show races when run with `-race`!

## 💡 Race Detector Usage

```bash
# Run tests with race detector
go test -race

# Run program with race detector
go run -race main.go

# Build with race detector
go build -race

# Run benchmarks with race detector
go test -race -bench=.
```

## 🎯 Common Race Patterns

### Unprotected shared variable
```go
// BAD: Race condition
var counter int
go func() { counter++ }()
go func() { counter++ }()
```

### Fixed with Mutex
```go
// GOOD: Protected with mutex
var mu sync.Mutex
var counter int
go func() {
    mu.Lock()
    counter++
    mu.Unlock()
}()
```

### Fixed with Channel
```go
// GOOD: Using channels
ch := make(chan int)
go func() { ch <- 1 }()
value := <-ch
```

## 🔍 What the Race Detector Shows

```
WARNING: DATA RACE
Write at 0x00c000018090 by goroutine 7:
  main.(*Counter).Increment()
      /path/to/main.go:15 +0x44

Previous read at 0x00c000018090 by goroutine 6:
  main.(*Counter).Value()
      /path/to/main.go:19 +0x34
```

## 📚 Resources

- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [Data Race Patterns](https://go.dev/ref/mem)
- [Sync Package](https://pkg.go.dev/sync)

## ✨ Synchronization Techniques

1. **Mutex**: Exclusive access (only one goroutine)
2. **RWMutex**: Multiple readers OR one writer
3. **Channels**: Communicate to share, don't share to communicate
4. **Atomic**: For simple counters (`sync/atomic`)
5. **sync.Once**: For one-time initialization

---

**Ready?** Open `main.go` and start coding! Run `go test -race -v` when you're done.
