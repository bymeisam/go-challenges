package main

import "sync"

type Database struct {
	Connection string
}

var instance *Database
var once sync.Once

func GetInstance() *Database {
	once.Do(func() {
		instance = &Database{Connection: "connected"}
	})
	return instance
}

func main() {}
