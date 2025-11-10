package main

import "strings"

func ContainsWord(text, word string) bool {
	return strings.Contains(text, word)
}

func SplitSentence(sentence string) []string {
	return strings.Split(sentence, " ")
}

func JoinWords(words []string) string {
	return strings.Join(words, " ")
}

func TrimSpaces(text string) string {
	return strings.TrimSpace(text)
}

func ReplaceAll(text, old, new string) string {
	return strings.ReplaceAll(text, old, new)
}

func ToUpperCase(text string) string {
	return strings.ToUpper(text)
}

func main() {}
