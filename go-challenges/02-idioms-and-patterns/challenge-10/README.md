# Challenge 10: Named Returns

**Difficulty:** ⭐⭐ Medium | **Time:** 15 min

## 🎯 Learning Goals
Learn Go's named return values - a unique feature for clearer code

## 🔨 Tasks
1. `Divide(a, b int) (result int, remainder int)` - named returns
2. `ReadConfig() (host string, port int, err error)` - return defaults on error
3. `ProcessData(data []int) (sum, count int)` - use naked return

## 💡 Key Concept
Named returns are pre-declared and can use "naked return" (`return` without values).

```bash
go test -v
```
