package main

import "regexp"

func MatchEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func FindAllNumbers(text string) []string {
	re := regexp.MustCompile(`\d+`)
	return re.FindAllString(text, -1)
}

func ReplaceDigits(text string) string {
	re := regexp.MustCompile(`\d`)
	return re.ReplaceAllString(text, "X")
}

func main() {}
