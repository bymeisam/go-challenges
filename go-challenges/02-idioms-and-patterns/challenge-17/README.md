# Challenge 17: Empty Interface vs Generics

**Difficulty:** ⭐⭐⭐ Hard | **Time:** 20 min

Learn when to use `interface{}` (any) vs generics (Go 1.18+).

## Tasks
1. `PrintAny(v interface{})` - accept any type
2. `GenericMax[T constraints.Ordered](a, b T) T` - generic function
3. `Contains[T comparable](slice []T, item T) bool` - generic slice search

```bash
go test -v
```
