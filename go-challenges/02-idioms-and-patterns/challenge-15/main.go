package main

import "errors"

type Database struct {
	ConnString string
}

func NewDatabase(conn string) (*Database, error) {
	if conn == "" {
		return nil, errors.New("connection string cannot be empty")
	}
	return &Database{ConnString: conn}, nil
}

func MustNewDatabase(conn string) *Database {
	db, err := NewDatabase(conn)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {}
