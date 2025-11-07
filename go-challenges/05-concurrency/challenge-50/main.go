package main

func SendOnly(ch chan<- int) {
	ch <- 42
}

func ReceiveOnly(ch <-chan int) int {
	return <-ch
}

func DirectionalChannels() int {
	ch := make(chan int, 1)
	SendOnly(ch)
	return ReceiveOnly(ch)
}

func main() {}
