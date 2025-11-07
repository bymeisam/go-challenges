package main

type Mover interface {
	Move() string
}

type Speaker interface {
	Speak() string
}

type Walker struct{}

func (w Walker) Move() string {
	return "walking"
}

type Talker struct{}

func (t Talker) Speak() string {
	return "talking"
}

type Dog struct {
	Walker
	Talker
}

type Car struct {
	Walker // Cars move but don't speak
}

func DescribeMover(m Mover) string {
	return "Moving by " + m.Move()
}

func main() {}
