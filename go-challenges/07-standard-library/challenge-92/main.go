package main

import (
	"bytes"
	"log"
)

func LogMessage(buf *bytes.Buffer, message string) {
	logger := log.New(buf, "INFO: ", log.Ldate|log.Ltime)
	logger.Println(message)
}

func LogError(buf *bytes.Buffer, err error) {
	logger := log.New(buf, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(err)
}

func main() {}
