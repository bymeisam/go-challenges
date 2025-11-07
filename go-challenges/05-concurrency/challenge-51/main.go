package main

import "time"

func SelectFromMultiple(ch1, ch2 chan string) string {
	select {
	case msg := <-ch1:
		return "ch1: " + msg
	case msg := <-ch2:
		return "ch2: " + msg
	}
}

func SelectWithTimeout(ch chan string, timeout time.Duration) (string, bool) {
	select {
	case msg := <-ch:
		return msg, true
	case <-time.After(timeout):
		return "", false
	}
}

func main() {}
