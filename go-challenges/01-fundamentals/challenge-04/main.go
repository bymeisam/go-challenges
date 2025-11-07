package main

import "fmt"

// TODO: Define a Person struct with Name (string) and Age (int) fields

// NewPerson creates and returns a new Person
func NewPerson(name string, age int) Person {
	// TODO: Create and return a Person struct
	return Person{}
}

// UpdateAge updates the age of a person (note the pointer receiver)
func UpdateAge(p *Person, newAge int) {
	// TODO: Update p.Age to newAge
}

// GetInfo returns a formatted string with person info
func GetInfo(p Person) string {
	// TODO: Return formatted string like "Name: John, Age: 30"
	return fmt.Sprintf("Name: %s, Age: %d", p.Name, p.Age)
}

func main() {}
