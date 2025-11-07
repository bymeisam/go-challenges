package main

import "fmt"

func Describe(v interface{}) string {
	switch val := v.(type) {
	case int:
		return fmt.Sprintf("Integer: %d", val)
	case string:
		return fmt.Sprintf("String: %s", val)
	case bool:
		return fmt.Sprintf("Boolean: %t", val)
	default:
		return "Unknown type"
	}
}

func main() {}
