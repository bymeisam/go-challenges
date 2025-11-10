package main

type PaymentStrategy interface {
	Pay(amount int) string
}

type CreditCard struct {
	Number string
}

func (c CreditCard) Pay(amount int) string {
	return "Paid with credit card"
}

type PayPal struct {
	Email string
}

func (p PayPal) Pay(amount int) string {
	return "Paid with PayPal"
}

type ShoppingCart struct {
	strategy PaymentStrategy
}

func (s *ShoppingCart) SetPaymentStrategy(strategy PaymentStrategy) {
	s.strategy = strategy
}

func (s *ShoppingCart) Checkout(amount int) string {
	return s.strategy.Pay(amount)
}

func main() {}
