package main

import (
	"bufio"
	"io"
	"strings"
)

func ReadLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	return lines, scanner.Err()
}

func CopyData(src io.Reader, dst io.Writer) (int64, error) {
	return io.Copy(dst, src)
}

func main() {}
