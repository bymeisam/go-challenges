package main

import (
	"strings"
	"testing"
)

func TestXML(t *testing.T) {
	book := Book{Title: "Go Programming", Author: "John Doe"}
	
	data, err := MarshalXML(book)
	if err != nil {
		t.Fatal("MarshalXML failed")
	}
	
	if !strings.Contains(string(data), "Go Programming") {
		t.Error("XML should contain title")
	}
	
	decoded, err := UnmarshalXML(data)
	if err != nil || decoded.Title != "Go Programming" {
		t.Error("UnmarshalXML failed")
	}
	
	t.Log("✓ encoding/xml works!")
}
