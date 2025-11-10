package main

import (
	"os"
	"path/filepath"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

func GetExtension(filename string) string {
	return filepath.Ext(filename)
}

func GetBasename(path string) string {
	return filepath.Base(path)
}

func main() {}
