# Challenge 33: Empty Interface

**Difficulty:** ⭐⭐ Medium | **Time:** 20 min

Master `interface{}` (or `any` in Go 1.18+) - can hold any type.

## Tasks
1. `Accept(v interface{})` - accept any type
2. `GetType(v interface{}) string` - return type name
3. `ConvertToInt(v interface{}) (int, bool)` - safe type conversion

```bash
go test -v
```
