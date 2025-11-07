# Challenge 11: Defer Statement

**Difficulty:** ⭐⭐ Medium | **Time:** 20 min

## 🎯 Learning Goals
Master defer - Go's unique cleanup mechanism (different from JS finally)

## 🔨 Tasks
1. `OpenAndProcess(filename string) error` - use defer to close file
2. `MeasureTime(operation string) func()` - return cleanup function with defer
3. `DeferOrder() []int` - understand LIFO defer execution

Go's `defer` schedules cleanup code that runs when function exits (like finally, but better!).

```bash
go test -v
```
