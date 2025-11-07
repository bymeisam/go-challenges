package main

func NonBlockingSend(ch chan int, value int) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}

func NonBlockingReceive(ch chan int) (int, bool) {
	select {
	case value := <-ch:
		return value, true
	default:
		return 0, false
	}
}

func main() {}
