# Solution for Challenge 10: Named Returns

```go
package main

func Divide(a, b int) (result int, remainder int) {
	result = a / b
	remainder = a % b
	return result, remainder  // Or just: return
}

func ReadConfig() (host string, port int, err error) {
	host = "localhost"
	port = 8080
	err = nil
	return  // Naked return
}

func ProcessData(data []int) (sum, count int) {
	for _, v := range data {
		sum += v
		count++
	}
	return  // Naked return
}
```

## Key Points
- Named returns are pre-declared in function signature
- Can use "naked return" (just `return`) - returns named values
- Makes code clearer, especially with multiple return values
- Useful for setting default values
