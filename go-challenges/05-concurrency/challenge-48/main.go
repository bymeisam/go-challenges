package main

func SendNumbers(ch chan int, numbers []int) {
	for _, num := range numbers {
		ch <- num
	}
	close(ch)
}

func ReceiveSum(ch chan int) int {
	sum := 0
	for num := range ch {
		sum += num
	}
	return sum
}

func Pipeline() int {
	ch := make(chan int)
	go SendNumbers(ch, []int{1, 2, 3, 4, 5})
	return ReceiveSum(ch)
}

func main() {}
