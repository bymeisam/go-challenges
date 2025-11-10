package main

import (
	"github.com/sirupsen/logrus"
)

func NewLogger(level string) *logrus.Logger {
	log := logrus.New()
	
	// Set formatter
	log.SetFormatter(&logrus.JSONFormatter{})
	
	// Set level
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)
	
	return log
}

func LogWithFields(log *logrus.Logger, level, message string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields(fields))
	
	switch level {
	case "debug":
		entry.Debug(message)
	case "info":
		entry.Info(message)
	case "warn":
		entry.Warn(message)
	case "error":
		entry.Error(message)
	default:
		entry.Info(message)
	}
}

func main() {
	log := NewLogger("info")
	
	log.Info("Application started")
	
	LogWithFields(log, "info", "User logged in", map[string]interface{}{
		"user_id": 123,
		"ip":      "192.168.1.1",
	})
	
	log.Warn("This is a warning")
	log.Error("This is an error")
}
