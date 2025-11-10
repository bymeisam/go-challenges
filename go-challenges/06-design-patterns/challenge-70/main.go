package main

type Coffee interface {
	Cost() int
	Description() string
}

type SimpleCoffee struct{}

func (s SimpleCoffee) Cost() int {
	return 5
}

func (s SimpleCoffee) Description() string {
	return "Simple coffee"
}

type MilkDecorator struct {
	coffee Coffee
}

func (m MilkDecorator) Cost() int {
	return m.coffee.Cost() + 2
}

func (m MilkDecorator) Description() string {
	return m.coffee.Description() + ", milk"
}

type SugarDecorator struct {
	coffee Coffee
}

func (s SugarDecorator) Cost() int {
	return s.coffee.Cost() + 1
}

func (s SugarDecorator) Description() string {
	return s.coffee.Description() + ", sugar"
}

func main() {}
