# Solution for Challenge 01: Variables and Types

Here's the complete solution with detailed explanations:

```go
package main

import (
	"strconv"
)

// DeclareVariables demonstrates zero values in Go.
// Return four zero-valued variables: int, string, bool, float64
func DeclareVariables() (int, string, bool, float64) {
	var i int
	var s string
	var b bool
	var f float64
	return i, s, b, f
}

// UseShortDeclaration uses := to declare and initialize a variable.
// Calculate and return the sum of x and y.
func UseShortDeclaration(x, y int) int {
	sum := x + y
	return sum
}

// TypeConversion converts a string to an integer.
// Return the converted integer and any error that occurs.
func TypeConversion(s string) (int, error) {
	return strconv.Atoi(s)
}

// ConstantsExample demonstrates constant declaration.
// Return the app name and version as constants.
func ConstantsExample() (string, int) {
	const appName = "GoChallenge"
	const version = 1
	return appName, version
}
```

## 📖 Detailed Explanation

### Function 1: DeclareVariables

```go
func DeclareVariables() (int, string, bool, float64) {
	var i int
	var s string
	var b bool
	var f float64
	return i, s, b, f
}
```

**What's happening:**
- We declare four variables using the `var` keyword without initialization
- Go automatically assigns **zero values**:
  - `int` → `0`
  - `string` → `""` (empty string)
  - `bool` → `false`
  - `float64` → `0.0`

**JS/TS Comparison:**
```javascript
// JavaScript
let i;        // undefined (not 0!)
let s;        // undefined (not "")
let b;        // undefined (not false)

// TypeScript
let i: number;  // undefined at runtime (but TS warns)
let s: string;  // undefined at runtime
```

In Go, there's **no `undefined`** or `null` for basic types. Variables always have a value!

### Function 2: UseShortDeclaration

```go
func UseShortDeclaration(x, y int) int {
	sum := x + y
	return sum
}
```

**What's happening:**
- The `:=` operator declares AND initializes a variable
- Go infers the type from the expression on the right
- `sum` is inferred to be `int` because `x + y` produces an `int`

**Alternate valid solutions:**
```go
// Option 1: Explicit type
var sum int = x + y
return sum

// Option 2: Type inference with var
var sum = x + y
return sum

// Option 3: Inline return (shortest)
return x + y
```

**JS/TS Comparison:**
```javascript
// JavaScript/TypeScript
const sum = x + y;  // Similar to Go's :=
return sum;
```

The main difference: `:=` can only be used **inside functions**, not at package level.

### Function 3: TypeConversion

```go
func TypeConversion(s string) (int, error) {
	return strconv.Atoi(s)
}
```

**What's happening:**
- `strconv.Atoi()` converts a string to an integer
- It returns **two values**: the result and an error
- If conversion succeeds, error is `nil`
- If it fails (e.g., "abc"), error contains the error details

**Why return both?**
Go uses explicit error handling instead of exceptions. The caller must check the error:

```go
result, err := TypeConversion("42")
if err != nil {
	// Handle error
	fmt.Println("Conversion failed:", err)
} else {
	// Use result
	fmt.Println("Converted:", result)
}
```

**JS/TS Comparison:**
```javascript
// JavaScript
const result = parseInt("42");     // 42
const invalid = parseInt("abc");   // NaN (no error!)

// TypeScript with error handling
function typeConversion(s: string): number {
	const num = parseInt(s);
	if (isNaN(num)) {
		throw new Error("Invalid number");
	}
	return num;
}
```

Go's approach is more explicit: errors are values you must handle.

### Function 4: ConstantsExample

```go
func ConstantsExample() (string, int) {
	const appName = "GoChallenge"
	const version = 1
	return appName, version
}
```

**What's happening:**
- `const` declares compile-time constants
- Constants must be literal values or expressions that can be evaluated at compile time
- You can't use `:=` for constants (constants use `const`, not `var` or `:=`)

**Alternate syntax:**
```go
// Group declaration (Go style)
const (
	appName = "GoChallenge"
	version = 1
)
```

**JS/TS Comparison:**
```javascript
// JavaScript/TypeScript
const appName = "GoChallenge";  // Similar syntax
const version = 1;

// But JS const only prevents reassignment:
const obj = {x: 1};
obj.x = 2;  // Allowed in JS!

// In Go, const values are truly immutable
```

**Important:** Go constants are more restrictive than JS constants:
- Must be compile-time values (no function calls)
- Cannot be objects, arrays, or slices
- Are truly immutable (unlike JS `const` objects)

## 🎓 Key Takeaways

1. **Zero values**: Go initializes variables automatically (no `undefined`)
2. **`:=` is for local variables**: Quick declaration + initialization
3. **Type inference**: Go can infer types, but it's still statically typed
4. **Explicit errors**: Go returns errors as values, not exceptions
5. **Constants are strict**: Must be compile-time values

## 🚀 Next Steps

Now that you understand Go's variables and types, try:
- Challenge 02 to learn about slices and arrays
- Experiment with type inference in `main()`
- Try the bonus challenge with multiple variable declarations

## 💡 Pro Tips

1. Use `:=` for local variables (it's idiomatic)
2. Use `var` when you need an explicit type or zero value
3. Always check errors returned from functions
4. Constants should be used for truly constant values (not config)

Great job completing this challenge! 🎉
