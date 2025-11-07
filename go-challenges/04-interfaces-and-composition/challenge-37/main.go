package main

import "fmt"

type Product struct {
	Name  string
	Price float64
}

func (p Product) String() string {
	return fmt.Sprintf("%s ($%.2f)", p.Name, p.Price)
}

type OutOfStockError struct {
	Product string
}

func (e OutOfStockError) Error() string {
	return fmt.Sprintf("product %s is out of stock", e.Product)
}

func main() {}
