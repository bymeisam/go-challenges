# Challenge 08: JSON Marshaling

**Difficulty:** ⭐⭐ Medium | **Time:** 20 min

## 🎯 Learning Goals
Learn to work with JSON (encoding/decoding) - essential for web APIs

## 🔨 Tasks
1. Define a `User` struct with JSON tags
2. `UserToJSON(user User) (string, error)` - marshal to JSON
3. `JSONToUser(jsonStr string) (User, error)` - unmarshal from JSON

## 💡 Key Concepts
- JSON tags: `` `json:"field_name"` ``
- `json.Marshal()` and `json.Unmarshal()`
- Similar to `JSON.stringify()` and `JSON.parse()` in JS

```bash
go test -v
```
