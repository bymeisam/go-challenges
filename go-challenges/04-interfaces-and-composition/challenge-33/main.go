package main

import "fmt"

func Accept(v interface{}) {
	fmt.Printf("Received: %v\n", v)
}

func GetType(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return "unknown"
	}
}

func ConvertToInt(v interface{}) (int, bool) {
	i, ok := v.(int)
	return i, ok
}

func main() {}
