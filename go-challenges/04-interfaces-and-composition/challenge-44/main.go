package main

// Bad: Large interface
// type Worker interface {
//     Work()
//     Eat()
//     Sleep()
// }

// Good: Segregated interfaces
type Worker interface {
	Work()
}

type Eater interface {
	Eat()
}

type Sleeper interface {
	Sleep()
}

type Robot struct {
	Name string
}

func (r Robot) Work() {
	// Robots work but don't eat or sleep
}

type Human struct {
	Name string
}

func (h Human) Work() {}
func (h Human) Eat() {}
func (h Human) Sleep() {}

func DoWork(w Worker) {
	w.Work()
}

func main() {}
