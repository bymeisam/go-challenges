# Challenge 130: Benchmarking

**Difficulty:** ⭐⭐ Medium
**Topic:** Testing & Performance
**Estimated Time:** 30-35 minutes

## 🎯 Learning Goals

- Understand benchmarking in Go
- Learn to write benchmark functions with `testing.B`
- Master using `b.N` for accurate measurements
- Practice analyzing benchmark results and comparing implementations

## 📝 Description

Benchmarks measure the performance of your code. Go's testing package provides built-in benchmarking:

1. **Benchmark functions**: Start with `Benchmark` and take `*testing.B`
2. **b.N**: The number of iterations (automatically determined)
3. **b.ResetTimer()**: Reset timer after setup
4. **Comparison**: Compare different implementations

Key benchmarking concepts:
- **ns/op**: Nanoseconds per operation
- **B/op**: Bytes allocated per operation
- **allocs/op**: Allocations per operation
- **-benchmem**: Show memory allocation stats

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. String concatenation functions

Implement three different approaches:
- `ConcatWithPlus(strs []string) string` - using `+` operator
- `ConcatWithBuilder(strs []string) string` - using `strings.Builder`
- `ConcatWithJoin(strs []string) string` - using `strings.Join`

### 2. Fibonacci implementations

Implement two approaches:
- `FibonacciRecursive(n int) int` - recursive implementation
- `FibonacciIterative(n int) int` - iterative implementation

### 3. Search functions

Implement linear and binary search:
- `LinearSearch(arr []int, target int) int` - return index or -1
- `BinarySearch(arr []int, target int) int` - assumes sorted array, return index or -1

### 4. Map operations

- `MapWithMake(size int) map[int]int` - create map with make and size hint
- `MapWithoutMake(size int) map[int]int` - create map without size hint

## 🧪 Testing and Benchmarking

Run tests:
```bash
go test -v
```

Run benchmarks:
```bash
# Run all benchmarks
go test -bench=.

# Run with memory stats
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkConcat

# Run with more iterations for accuracy
go test -bench=. -benchtime=10s
```

All tests and benchmarks must run! ✅

## 💡 Benchmark Pattern

```go
func BenchmarkExample(b *testing.B) {
    // Setup (not measured)
    data := setupTestData()

    // Reset timer after setup
    b.ResetTimer()

    // Run the function b.N times
    for i := 0; i < b.N; i++ {
        FunctionToTest(data)
    }
}
```

## 🎯 Reading Benchmark Results

```
BenchmarkConcatWithPlus-8        1000000    1234 ns/op    512 B/op    10 allocs/op
```

- `1000000`: Number of iterations (b.N)
- `1234 ns/op`: Nanoseconds per operation
- `512 B/op`: Bytes allocated per operation
- `10 allocs/op`: Memory allocations per operation
- `-8`: Number of CPU cores used (GOMAXPROCS)

## 📚 Resources

- [Go Testing: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Benchmarking in Go](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)

## ✨ Performance Tips

1. **Use strings.Builder**: For string concatenation
2. **Preallocate slices/maps**: When you know the size
3. **Avoid unnecessary allocations**: Reuse objects
4. **Choose right algorithm**: Binary search vs linear
5. **Iterative over recursive**: Often faster, less memory

---

**Ready?** Open `main.go` and start coding! Run `go test -bench=. -benchmem` when you're done.
