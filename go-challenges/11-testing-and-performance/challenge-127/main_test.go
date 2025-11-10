package main

import (
	"testing"
)

func TestCalculator(t *testing.T) {
	calc := Calculator{}

	t.Run("Addition", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected int
		}{
			{"positive", 2, 3, 5},
			{"negative", -1, -1, -2},
			{"zero", 0, 5, 5},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := calc.Add(tt.a, tt.b)
				if result != tt.expected {
					t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})

	t.Run("Subtraction", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected int
		}{
			{"positive", 5, 3, 2},
			{"negative result", 3, 5, -2},
			{"zero", 5, 5, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := calc.Subtract(tt.a, tt.b)
				if result != tt.expected {
					t.Errorf("Subtract(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})

	t.Run("Multiplication", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b     int
			expected int
		}{
			{"positive", 3, 4, 12},
			{"by zero", 5, 0, 0},
			{"negative", -2, 3, -6},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := calc.Multiply(tt.a, tt.b)
				if result != tt.expected {
					t.Errorf("Multiply(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
				}
			})
		}
	})

	t.Run("Division", func(t *testing.T) {
		t.Run("valid division", func(t *testing.T) {
			result, err := calc.Divide(10, 2)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != 5.0 {
				t.Errorf("Divide(10, 2) = %f; want 5.0", result)
			}
		})

		t.Run("division by zero", func(t *testing.T) {
			_, err := calc.Divide(10, 0)
			if err == nil {
				t.Error("expected error for division by zero, got nil")
			}
		})

		t.Run("float division", func(t *testing.T) {
			result, err := calc.Divide(7, 2)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != 3.5 {
				t.Errorf("Divide(7, 2) = %f; want 3.5", result)
			}
		})
	})

	t.Log("✓ All Calculator tests passed!")
}

func TestUserValidator(t *testing.T) {
	validator := UserValidator{}

	t.Run("ValidateUsername", func(t *testing.T) {
		t.Run("valid usernames", func(t *testing.T) {
			validNames := []string{"user123", "john", "alice2023"}
			for _, name := range validNames {
				t.Run(name, func(t *testing.T) {
					err := validator.ValidateUsername(name)
					if err != nil {
						t.Errorf("ValidateUsername(%q) returned error: %v", name, err)
					}
				})
			}
		})

		t.Run("invalid usernames", func(t *testing.T) {
			invalidNames := []string{
				"ab",                      // too short
				"verylongusername12345678", // too long
				"user@123",                // special char
				"user name",               // space
			}
			for _, name := range invalidNames {
				t.Run(name, func(t *testing.T) {
					err := validator.ValidateUsername(name)
					if err == nil {
						t.Errorf("ValidateUsername(%q) should return error", name)
					}
				})
			}
		})
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		t.Run("valid passwords", func(t *testing.T) {
			validPasswords := []string{"password1", "secret123", "mypass99"}
			for _, pwd := range validPasswords {
				t.Run(pwd, func(t *testing.T) {
					err := validator.ValidatePassword(pwd)
					if err != nil {
						t.Errorf("ValidatePassword(%q) returned error: %v", pwd, err)
					}
				})
			}
		})

		t.Run("invalid passwords", func(t *testing.T) {
			tests := []struct {
				name     string
				password string
			}{
				{"too short", "pass1"},
				{"no number", "password"},
				{"empty", ""},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					err := validator.ValidatePassword(tt.password)
					if err == nil {
						t.Errorf("ValidatePassword(%q) should return error", tt.password)
					}
				})
			}
		})
	})

	t.Run("ValidateAge", func(t *testing.T) {
		t.Run("valid ages", func(t *testing.T) {
			validAges := []int{18, 25, 100}
			for _, age := range validAges {
				err := validator.ValidateAge(age)
				if err != nil {
					t.Errorf("ValidateAge(%d) returned error: %v", age, err)
				}
			}
		})

		t.Run("invalid ages", func(t *testing.T) {
			invalidAges := []int{0, 10, 17}
			for _, age := range invalidAges {
				err := validator.ValidateAge(age)
				if err == nil {
					t.Errorf("ValidateAge(%d) should return error", age)
				}
			}
		})
	})

	t.Log("✓ All UserValidator tests passed!")
}

func TestStringProcessor(t *testing.T) {
	sp := StringProcessor{}

	t.Run("Reverse", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"hello", "olleh"},
			{"Go", "oG"},
			{"racecar", "racecar"},
			{"", ""},
			{"a", "a"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := sp.Reverse(tt.input)
				if result != tt.expected {
					t.Errorf("Reverse(%q) = %q; want %q", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("WordCount", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int
		}{
			{"simple", "hello world", 2},
			{"multiple spaces", "hello   world", 2},
			{"single word", "hello", 1},
			{"empty", "", 0},
			{"spaces only", "   ", 0},
			{"many words", "the quick brown fox", 4},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := sp.WordCount(tt.input)
				if result != tt.expected {
					t.Errorf("WordCount(%q) = %d; want %d", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("ToSnakeCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"Hello World", "hello_world"},
			{"helloWorld", "hello_world"},
			{"HelloWorld", "hello_world"},
			{"hello", "hello"},
			{"HTTP Server", "h_t_t_p_server"},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := sp.ToSnakeCase(tt.input)
				if result != tt.expected {
					t.Errorf("ToSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Log("✓ All StringProcessor tests passed!")
}
