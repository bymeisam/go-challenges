package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func FormatPerson(p Person) string {
	return fmt.Sprintf("%s is %d years old", p.Name, p.Age)
}

func FormatWithTypes(s string, n int, f float64) string {
	return fmt.Sprintf("String: %s, Int: %d, Float: %.2f", s, n, f)
}

func FormatStruct(p Person) string {
	return fmt.Sprintf("%+v", p)
}

func main() {}
