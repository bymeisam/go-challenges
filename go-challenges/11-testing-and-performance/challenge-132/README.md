# Challenge 132: Profiling

**Difficulty:** ⭐⭐⭐ Hard
**Topic:** Testing & Performance
**Estimated Time:** 35-40 minutes

## 🎯 Learning Goals

- Understand CPU and memory profiling in Go
- Learn to generate and analyze profile data
- Master using pprof for performance analysis
- Practice identifying performance bottlenecks

## 📝 Description

Profiling helps you understand where your program spends time and allocates memory. Go provides built-in profiling tools:

1. **CPU profiling**: Shows where CPU time is spent
2. **Memory profiling**: Shows memory allocation patterns
3. **Block profiling**: Shows goroutine blocking
4. **Mutex profiling**: Shows lock contention

The `pprof` tool analyzes profile data and shows:
- Hot spots (functions using most CPU/memory)
- Call graphs
- Flame graphs

## 🔨 Your Task

Implement the following in `main.go`:

### 1. `ProcessData(data []int) []int`

Process a slice of integers:
- Square each number
- Filter out odd numbers
- Sort the result

### 2. `GenerateReport(n int) string`

Generate a report with string concatenation:
- Create n lines of text
- Each line contains formatted data
- Return as single string

### 3. `FindPrimes(max int) []int`

Find all prime numbers up to max:
- Use Sieve of Eratosthenes algorithm
- Return slice of primes

### 4. `MatrixMultiply(a, b [][]int) [][]int`

Multiply two square matrices:
- Assume square matrices of same size
- Return result matrix

### 5. `ImageProcessor`

```go
type ImageProcessor struct {}
func (ip *ImageProcessor) ProcessImage(width, height int) [][]int
```
Create and process a 2D image (matrix of pixels).

## 🧪 Testing and Profiling

Run tests:
```bash
go test -v
```

Generate CPU profile:
```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=.

# View CPU profile
go tool pprof cpu.prof
# In pprof: top, list ProcessData, web

# Memory profile
go test -memprofile=mem.prof -bench=.

# View memory profile
go tool pprof mem.prof
# In pprof: top, list, alloc_space
```

Analyze profiles:
```bash
# Interactive mode
go tool pprof cpu.prof

# Web interface (requires graphviz)
go tool pprof -http=:8080 cpu.prof

# Text output
go tool pprof -text cpu.prof

# Show top functions
go tool pprof -top cpu.prof
```

## 💡 Profiling Commands

In pprof interactive mode:
```
(pprof) top          # Show top functions
(pprof) top10        # Show top 10
(pprof) list Func    # Show annotated source
(pprof) web          # Open in browser (needs graphviz)
(pprof) pdf          # Generate PDF (needs graphviz)
(pprof) help         # Show all commands
```

## 🎯 Profile Data in Tests

You can also enable profiling programmatically:

```go
import (
    "os"
    "runtime/pprof"
)

func BenchmarkExample(b *testing.B) {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Your benchmark code
}
```

## 🔍 What to Look For

CPU Profile:
- Functions with high `flat` time (self time)
- Functions with high `cum` time (cumulative time)
- Unexpected allocations

Memory Profile:
- Large allocations
- Frequent small allocations
- Memory leaks (growing over time)

## 📚 Resources

- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Runtime/pprof Package](https://pkg.go.dev/runtime/pprof)
- [Go Diagnostics](https://go.dev/doc/diagnostics)
- [Flame Graphs](https://www.brendangregg.com/flamegraphs.html)

## ✨ Performance Optimization Tips

1. **Profile first**: Don't optimize without profiling
2. **Focus on hot spots**: 80/20 rule applies
3. **Measure changes**: Profile before and after
4. **Reduce allocations**: Reuse objects, use sync.Pool
5. **Algorithm matters**: O(n²) vs O(n log n) makes a difference

## 🎨 Advanced Profiling

```bash
# Block profiling (goroutine blocking)
go test -blockprofile=block.prof -bench=.

# Mutex profiling (lock contention)
go test -mutexprofile=mutex.prof -bench=.

# Trace (detailed execution trace)
go test -trace=trace.out -bench=.
go tool trace trace.out
```

---

**Ready?** Open `main.go` and start coding! Run `go test -cpuprofile=cpu.prof -bench=.` when you're done.
