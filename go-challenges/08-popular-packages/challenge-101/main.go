package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	APIKey     string
	Debug      string
}

func LoadEnv(filename string) error {
	return godotenv.Load(filename)
}

func LoadEnvWithDefault() error {
	return godotenv.Load()
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func LoadConfig() *Config {
	return &Config{
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBPort:     GetEnv("DB_PORT", "5432"),
		DBUser:     GetEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		APIKey:     os.Getenv("API_KEY"),
		Debug:      GetEnv("DEBUG", "false"),
	}
}

func main() {
	if err := LoadEnvWithDefault(); err != nil {
		fmt.Println("No .env file found, using defaults")
	}
	
	config := LoadConfig()
	fmt.Printf("DB Host: %s\n", config.DBHost)
	fmt.Printf("DB Port: %s\n", config.DBPort)
}
