package main

import "encoding/json"

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func MarshalUser(user User) ([]byte, error) {
	return json.Marshal(user)
}

func UnmarshalUser(data []byte) (User, error) {
	var user User
	err := json.Unmarshal(data, &user)
	return user, err
}

func MarshalSlice(users []User) ([]byte, error) {
	return json.Marshal(users)
}

func main() {}
