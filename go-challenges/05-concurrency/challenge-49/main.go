package main

func BufferedChannel(capacity int) chan string {
	return make(chan string, capacity)
}

func TrySend(ch chan string, value string) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}

func main() {}
