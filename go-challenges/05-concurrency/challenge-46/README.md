# Challenge 46: Basic Goroutines

**Difficulty:** ⭐⭐ Medium | **Topic:** Concurrency | **Time:** 25 min

## 🎯 Learning Goals
- Understand goroutines (lightweight threads)
- Learn the `go` keyword
- Compare with JavaScript async/await and Web Workers

## 📝 Description

Goroutines are Go's approach to concurrent execution. They're like super-lightweight threads - you can easily run thousands of them!

```go
go myFunction()  // Runs concurrently!
```

**Key difference from JS:**
- JavaScript: Single-threaded with async/await (cooperative)
- Go: True parallelism with goroutines (can use multiple CPU cores)

## 🔨 Your Task

### 1. `RunConcurrently(n int)`
Launch `n` goroutines that each print a number.

### 2. `ConcurrentSum(numbers []int) int`
Calculate sum using multiple goroutines. Use channels to collect results.

### 3. `Race()` 
Demonstrate a race condition (we'll fix it in the next challenge).

## 💡 JS vs Go

```javascript
// JavaScript - async (still single-threaded!)
async function fetchData() {
    const result = await fetch(url);
    return result;
}

// Go - true parallelism!
go func() {
    result := fetchData()
}()
```

## 🧪 Testing
```bash
cd go-challenges/05-concurrency/challenge-46
go test -v
```

---
**Ready?** Open `main.go` and start coding!
