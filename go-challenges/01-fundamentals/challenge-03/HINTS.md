# Hints for Challenge 03: Maps

## Hint 1: CreateMap
<details>
<summary>Click to reveal</summary>

```go
// Map literal
m := map[string]int{
	"one":   1,
	"two":   2,
	"three": 3,
}

// Or using make
m := make(map[string]int)
m["one"] = 1
// ...
```
</details>

## Hint 2: AddToMap
<details>
<summary>Click to reveal</summary>

```go
m[key] = value  // That's it!
```
Maps are reference types, so modifications are visible to the caller.
</details>

## Hint 3: GetFromMap
<details>
<summary>Click to reveal</summary>

Use the "comma-ok" idiom:
```go
value, ok := m[key]
return value, ok
```
If key doesn't exist: `value` is zero value, `ok` is false.
</details>

## Hint 4: DeleteFromMap
<details>
<summary>Click to reveal</summary>

```go
delete(m, key)  // Built-in function
```
</details>

## Hint 5: MapKeys
<details>
<summary>Click to reveal</summary>

```go
keys := make([]string, 0, len(m))
for key := range m {
	keys = append(keys, key)
}
return keys
```
Note: Map iteration order is random in Go!
</details>
