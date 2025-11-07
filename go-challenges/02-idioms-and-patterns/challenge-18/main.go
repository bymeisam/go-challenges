package main

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return "Some sound"
}

type Dog struct {
	Animal  // Embedding
	Breed string
}

func (d Dog) Speak() string {
	return "Woof!"
}

func main() {}
