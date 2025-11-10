package main

import (
	"strings"
	"testing"
)

func TestURL(t *testing.T) {
	rawURL := "https://example.com/path?key=value"
	
	u, err := ParseURL(rawURL)
	if err != nil || u.Scheme != "https" {
		t.Error("ParseURL failed")
	}
	
	host, err := GetHost(rawURL)
	if err != nil || host != "example.com" {
		t.Error("GetHost failed")
	}
	
	result, err := AddQueryParam("https://example.com", "foo", "bar")
	if err != nil || !strings.Contains(result, "foo=bar") {
		t.Errorf("AddQueryParam failed: %s", result)
	}
	
	t.Log("✓ net/url works!")
}
