package main

// DeclareVariables demonstrates zero values in Go.
// Return four zero-valued variables: int, string, bool, float64
func DeclareVariables() (int, string, bool, float64) {
	// TODO: Declare four variables without initialization
	// and return them
	return 0, "", false, 0.0
}

// UseShortDeclaration uses := to declare and initialize a variable.
// Calculate and return the sum of x and y.
func UseShortDeclaration(x, y int) int {
	// TODO: Use := to declare a variable 'sum'
	// that holds x + y, then return it
	return 0
}

// TypeConversion converts a string to an integer.
// Return the converted integer and any error that occurs.
// HINT: You'll need to import "strconv" package
func TypeConversion(s string) (int, error) {
	// TODO: Use strconv.Atoi to convert the string to int
	// Return both the result and the error
	return 0, nil
}

// ConstantsExample demonstrates constant declaration.
// Return the app name and version as constants.
func ConstantsExample() (string, int) {
	// TODO: Declare two constants:
	// - appName = "GoChallenge"
	// - version = 1
	// Then return them
	return "", 0
}

func main() {
	// You can test your functions here
	// This won't be graded, but it's useful for manual testing
}
