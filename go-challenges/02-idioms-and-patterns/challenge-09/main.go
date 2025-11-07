package main

// Counter represents a simple counter
type Counter struct {
	Count int
}

// SafeCounter creates a new Counter leveraging zero values
func SafeCounter() *Counter {
	// TODO: Return a new Counter without explicit initialization
	// The zero value of int (0) should be sufficient
	return nil
}

// IsZeroValue checks if a value is its zero value
func IsZeroValue(v interface{}) bool {
	// TODO: Use type assertion to check if v is zero value
	// Handle int, string, and bool types
	return false
}

// InitializeSlice returns a non-nil empty slice
func InitializeSlice() []int {
	// TODO: Return an empty slice that is NOT nil
	// Hint: Use make() or slice literal
	return nil
}

// SafeStringOperation safely handles potentially nil string pointers
func SafeStringOperation(s *string) string {
	// TODO: If s is nil, return "default"
	// Otherwise, return the value pointed to by s
	return ""
}

func main() {}
