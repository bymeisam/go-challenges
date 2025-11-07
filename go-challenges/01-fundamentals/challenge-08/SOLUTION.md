# Solution for Challenge 08: JSON Marshaling

```go
package main

import "encoding/json"

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func UserToJSON(user User) (string, error) {
	bytes, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func JSONToUser(jsonStr string) (User, error) {
	var user User
	err := json.Unmarshal([]byte(jsonStr), &user)
	return user, err
}
```

## Key Points
- JSON tags control field names in JSON: `` `json:"field_name"` ``
- `json.Marshal()` converts struct → JSON bytes
- `json.Unmarshal()` converts JSON bytes → struct
- Similar to JS `JSON.stringify()` and `JSON.parse()`

## Common JSON Tags
```go
type Example struct {
    Field1 string `json:"field1"`           // Rename
    Field2 string `json:"field2,omitempty"` // Omit if empty
    Field3 string `json:"-"`                // Skip this field
}
```
