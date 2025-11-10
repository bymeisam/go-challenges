package main

import (
	"strings"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := GetCurrentTime()
	if now.IsZero() {
		t.Error("GetCurrentTime failed")
	}
	
	later := AddDuration(now, 2)
	if later.Sub(now) != 2*time.Hour {
		t.Error("AddDuration failed")
	}
	
	formatted := FormatTime(now)
	if !strings.Contains(formatted, "-") {
		t.Error("FormatTime failed")
	}
	
	parsed, err := ParseTime("2024-01-01")
	if err != nil || parsed.Year() != 2024 {
		t.Error("ParseTime failed")
	}
	
	diff := TimeDifference(later, now)
	if diff != 2*time.Hour {
		t.Error("TimeDifference failed")
	}
	
	t.Log("✓ time package works!")
}
