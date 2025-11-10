package main

import "strconv"

func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func main() {}
