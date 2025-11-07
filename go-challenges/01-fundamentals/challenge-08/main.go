package main

import "encoding/json"

// TODO: Define User struct with Name, Email, Age fields
// Add JSON tags like: `json:"name"`

// UserToJSON converts a User to JSON string
func UserToJSON(user User) (string, error) {
	// TODO: Use json.Marshal to convert user to JSON bytes
	// Convert bytes to string
	return "", nil
}

// JSONToUser converts JSON string to User
func JSONToUser(jsonStr string) (User, error) {
	// TODO: Use json.Unmarshal to convert JSON bytes to User
	var user User
	return user, nil
}

func main() {}
