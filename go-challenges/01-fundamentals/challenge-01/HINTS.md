# Hints for Challenge 01: Variables and Types

Need a little help? Here are some hints to guide you:

## Hint 1: DeclareVariables

<details>
<summary>Click to reveal hint</summary>

In Go, you can declare variables without initialization and they automatically get zero values:

```go
var myInt int        // automatically 0
var myString string  // automatically ""
var myBool bool      // automatically false
var myFloat float64  // automatically 0.0
```

Just declare four variables with the correct types and return them!

</details>

## Hint 2: UseShortDeclaration

<details>
<summary>Click to reveal hint</summary>

The short declaration operator `:=` is used like this:

```go
variableName := value
```

For example:
```go
sum := x + y
```

This is similar to TypeScript's `let sum = x + y`, but Go infers the type automatically.

</details>

## Hint 3: TypeConversion

<details>
<summary>Click to reveal hint</summary>

The `strconv.Atoi()` function converts a string to an integer:

```go
result, err := strconv.Atoi(s)
```

It returns two values:
1. The converted integer
2. An error (if conversion fails, otherwise `nil`)

Just return both values directly!

**Important:** In Go, error handling is explicit. Unlike JavaScript where `parseInt("abc")` returns `NaN`, Go returns an error that you must handle.

</details>

## Hint 4: ConstantsExample

<details>
<summary>Click to reveal hint</summary>

Constants in Go are declared with the `const` keyword:

```go
const myConstant = value
```

For example:
```go
const appName = "GoChallenge"
const version = 1
```

Then just return them like regular variables.

**Note:** Go constants must be compile-time constants (literal values), unlike JavaScript's `const` which just prevents reassignment.

</details>

## Still Stuck?

Check out `SOLUTION.md` for a complete working solution with explanations!

## JS/TS Developer Tips

- **No `undefined`**: Go uses zero values instead
- **Explicit types**: Go is statically typed (like TypeScript)
- **`:=` is your friend**: Use it for local variables (similar to `let` in JS)
- **Error handling**: Go returns errors explicitly, not exceptions
