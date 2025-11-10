package main

import "sync"

type Config struct {
	once sync.Once
	data string
}

func (c *Config) Load() {
	c.once.Do(func() {
		c.data = "loaded"
	})
}

func (c *Config) Get() string {
	return c.data
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.Load()
	})
	return instance
}

func main() {}
