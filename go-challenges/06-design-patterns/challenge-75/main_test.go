package main

import "testing"

func TestTemplateMethod(t *testing.T) {
	csv := CSVProcessor{}
	base := &BaseProcessor{processor: csv}
	
	result := base.Execute()
	if result != "Saved Processed CSV data" {
		t.Errorf("Template method failed: %s", result)
	}
	
	t.Log("✓ Template method works!")
}
