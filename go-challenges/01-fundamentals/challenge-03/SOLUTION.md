# Solution for Challenge 03: Maps

```go
package main

func CreateMap() map[string]int {
	return map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
}

func AddToMap(m map[string]int, key string, value int) {
	m[key] = value
}

func GetFromMap(m map[string]int, key string) (int, bool) {
	value, ok := m[key]
	return value, ok
}

func DeleteFromMap(m map[string]int, key string) {
	delete(m, key)
}

func MapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
```

## Key Points

1. **Map literal**: `map[KeyType]ValueType{k1: v1, k2: v2}`
2. **Comma-ok idiom**: `v, ok := m[k]` - check existence
3. **Delete**: `delete(m, key)` built-in function
4. **Iteration**: `for k, v := range m` (random order!)
5. **Reference type**: Changes to map are visible to all references

## JS Comparison

```javascript
// JavaScript Object
const obj = {one: 1, two: 2};
obj["test"] = 42;           // Add
const val = obj["test"];    // Get (undefined if missing)
delete obj["test"];         // Delete
"test" in obj;              // Check existence

// ES6 Map
const map = new Map([["one", 1], ["two", 2]]);
map.set("test", 42);        // Add
map.get("test");            // Get
map.has("test");            // Check
map.delete("test");         // Delete
```

## Pro Tips
- Maps are NOT safe for concurrent access (use sync.Map or mutex)
- Iteration order is intentionally randomized
- Zero value of a map is `nil` (can't add to nil map)
- Preallocate capacity: `make(map[K]V, capacity)`
