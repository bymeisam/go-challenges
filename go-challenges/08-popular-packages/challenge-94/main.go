package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Debug    bool
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func InitConfig() (*viper.Viper, error) {
	v := viper.New()
	
	// Set defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("debug", false)
	
	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	
	return v, nil
}

func GetConfig(v *viper.Viper) (*Config, error) {
	var config Config
	
	config.Server.Host = v.GetString("server.host")
	config.Server.Port = v.GetInt("server.port")
	config.Database.Host = v.GetString("database.host")
	config.Database.Port = v.GetInt("database.port")
	config.Database.Username = v.GetString("database.username")
	config.Database.Password = v.GetString("database.password")
	config.Debug = v.GetBool("debug")
	
	return &config, nil
}

func main() {
	v, err := InitConfig()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		return
	}
	
	config, err := GetConfig(v)
	if err != nil {
		fmt.Printf("Error getting config: %v\n", err)
		return
	}
	
	fmt.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
}
