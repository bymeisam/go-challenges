package main

type Logger interface {
	Log(message string)
}

type FileLogger struct {
	Filename string
	Messages []string
}

func (f *FileLogger) Log(message string) {
	f.Messages = append(f.Messages, message)
}

type ConsoleLogger struct {
	Messages []string
}

func (c *ConsoleLogger) Log(message string) {
	c.Messages = append(c.Messages, message)
}

func LogMessage(logger Logger, msg string) {
	logger.Log(msg)
}

func main() {}
