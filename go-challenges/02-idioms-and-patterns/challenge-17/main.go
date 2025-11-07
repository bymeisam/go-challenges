package main

import (
	"fmt"
	"golang.org/x/exp/constraints"
)

func PrintAny(v interface{}) {
	fmt.Println(v)
}

func GenericMax[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func main() {}
