# Challenge 01: Variables and Types

**Difficulty:** ⭐ Easy
**Topic:** Fundamentals
**Estimated Time:** 15-20 minutes

## 🎯 Learning Goals

- Understand Go's type system and type inference
- Learn the difference between `var`, `:=`, and `const`
- Practice working with basic Go types
- Understand zero values (different from JavaScript's `undefined`)

## 📝 Description

As a JavaScript/TypeScript developer, you're used to `let`, `const`, and dynamic typing (or TypeScript's type annotations). Go has a stricter type system with some interesting differences:

1. **Zero values**: Variables without explicit initialization get "zero values" (not `undefined`!)
2. **Type inference**: Go can infer types with `:=` (similar to TS's `let x = ...`)
3. **No implicit coercion**: Types don't convert automatically like in JS

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. `DeclareVariables() (int, string, bool, float64)`

Declare and return four variables with these zero values:
- An integer
- A string
- A boolean
- A float64

**Hint:** In Go, zero values are: `0` for numbers, `""` for strings, `false` for bools.

### 2. `UseShortDeclaration(x, y int) int`

Use the short declaration operator (`:=`) to create a local variable `sum` and return it.

### 3. `TypeConversion(s string) (int, error)`

Convert a string to an integer. In JavaScript, you might use `parseInt()` or `Number()`. In Go, use the `strconv` package.

**Hint:** Use `strconv.Atoi(s)` which returns `(int, error)`.

### 4. `ConstantsExample() (string, int)`

Declare two constants inside the function:
- `appName` = "GoChallenge"
- `version` = 1

Return both.

## 🧪 Testing

Run the tests to verify your solution:

```bash
cd go-challenges/01-fundamentals/challenge-01
go test -v
```

All tests must pass! ✅

## 💡 Key Differences from JavaScript/TypeScript

| Concept | JavaScript/TypeScript | Go |
|---------|----------------------|-----|
| Uninitialized vars | `undefined` | Zero value (0, "", false, nil) |
| Declaration | `let x = 5` | `var x int = 5` or `x := 5` |
| Const | `const x = 5` | `const x = 5` (must be compile-time constant) |
| Type conversion | `Number("42")` | `strconv.Atoi("42")` (returns error) |
| Dynamic typing | Yes (in JS) | No (always statically typed) |

## 📚 Resources

- [Go by Example: Variables](https://gobyexample.com/variables)
- [Tour of Go: Basic Types](https://go.dev/tour/basics/11)
- [Effective Go: Names](https://go.dev/doc/effective_go#names)

## ✨ Bonus Challenge (Optional)

Try to declare multiple variables in one line:
```go
var x, y, z int = 1, 2, 3
```

Can you do the same with `:=`?

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
