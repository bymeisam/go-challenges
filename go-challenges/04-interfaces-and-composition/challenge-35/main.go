package main

import "io"

type StringReader struct {
	data string
	pos  int
}

func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

type ByteWriter struct {
	data []byte
}

func (w *ByteWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func main() {}
