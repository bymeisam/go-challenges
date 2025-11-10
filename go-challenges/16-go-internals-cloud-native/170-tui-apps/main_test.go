package main

import (
	"testing"
	"time"
)

// TestTextInput tests text input component
func TestTextInputBasic(t *testing.T) {
	input := NewTextInput("Name")

	if input.Placeholder != "Name" {
		t.Errorf("Expected placeholder 'Name', got %s", input.Placeholder)
	}

	if input.Focus {
		t.Errorf("Input should not be focused by default")
	}
}

func TestTextInputUpdate(t *testing.T) {
	input := NewTextInput("Test")
	input.SetFocus(true)

	input.Update(KeyMsg{Type: KeyType(rune('A')), Rune: 'A'})

	input.mu.RLock()
	if input.Value != "A" {
		t.Errorf("Expected 'A', got %s", input.Value)
	}
	input.mu.RUnlock()
}

func TestTextInputView(t *testing.T) {
	input := NewTextInput("Test")
	input.Value = "Hello"

	view := input.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}
}

func TestTextInputMaxLength(t *testing.T) {
	input := NewTextInput("Test")
	input.MaxLength = 5
	input.SetFocus(true)

	for _, ch := range "HelloWorld" {
		input.Update(KeyMsg{Type: KeyType(ch), Rune: ch})
	}

	input.mu.RLock()
	if len(input.Value) > input.MaxLength {
		t.Errorf("Value exceeded max length")
	}
	input.mu.RUnlock()
}

// TestList tests list component
func TestListBasic(t *testing.T) {
	list := NewList()

	if len(list.Items) != 0 {
		t.Errorf("Expected empty list initially")
	}
}

func TestListAddItem(t *testing.T) {
	list := NewList()
	list.AddItem("Item 1", "val1")
	list.AddItem("Item 2", "val2")

	list.mu.RLock()
	if len(list.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.Items))
	}
	list.mu.RUnlock()
}

func TestListNavigation(t *testing.T) {
	list := NewList()
	list.AddItem("First", nil)
	list.AddItem("Second", nil)
	list.AddItem("Third", nil)
	list.SetFocus(true)

	list.Update(KeyMsg{Type: KeyDown})

	list.mu.RLock()
	if list.Cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", list.Cursor)
	}
	list.mu.RUnlock()
}

func TestListView(t *testing.T) {
	list := NewList()
	list.AddItem("Item 1", nil)

	view := list.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}
}

func TestListGetSelectedValue(t *testing.T) {
	list := NewList()
	list.AddItem("Item 1", "value1")
	list.AddItem("Item 2", "value2")

	value := list.GetSelectedValue()
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
}

// TestTable tests table component
func TestTableBasic(t *testing.T) {
	table := NewTable([]string{"Name", "Age"})

	table.mu.RLock()
	if len(table.Headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(table.Headers))
	}
	table.mu.RUnlock()
}

func TestTableAddRow(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2"})
	table.AddRow([]string{"a", "b"})
	table.AddRow([]string{"c", "d"})

	table.mu.RLock()
	if len(table.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.Rows))
	}
	table.mu.RUnlock()
}

func TestTableView(t *testing.T) {
	table := NewTable([]string{"Name"})
	table.AddRow([]string{"Alice"})

	view := table.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}
}

func TestTableNavigation(t *testing.T) {
	table := NewTable([]string{"Col"})
	table.AddRow([]string{"Row1"})
	table.AddRow([]string{"Row2"})
	table.SetFocus(true)

	table.Update(KeyMsg{Type: KeyDown})

	table.mu.RLock()
	if table.CursorRow != 1 {
		t.Errorf("Expected cursor at 1, got %d", table.CursorRow)
	}
	table.mu.RUnlock()
}

// TestProgressBar tests progress bar
func TestProgressBarBasic(t *testing.T) {
	pb := NewProgressBar(100, "Test")

	if pb.Total != 100 {
		t.Errorf("Expected total 100, got %d", pb.Total)
	}
}

func TestProgressBarUpdate(t *testing.T) {
	pb := NewProgressBar(100, "Test")
	pb.Update(50)

	current := pb.Current
	if current != 50 {
		t.Errorf("Expected current 50, got %d", current)
	}
}

func TestProgressBarView(t *testing.T) {
	pb := NewProgressBar(100, "Test")
	pb.Update(50)

	view := pb.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}

	if !contains(view, "50%") {
		t.Errorf("Expected percentage in view")
	}
}

func TestProgressBarETA(t *testing.T) {
	pb := NewProgressBar(100, "Test")
	pb.Update(50)

	time.Sleep(100 * time.Millisecond)

	view := pb.View()
	if !contains(view, "ETA") {
		t.Errorf("Expected ETA in view")
	}
}

// TestForm tests form component
func TestFormBasic(t *testing.T) {
	form := NewForm()

	if len(form.Fields) != 0 {
		t.Errorf("Expected empty form initially")
	}
}

func TestFormAddField(t *testing.T) {
	form := NewForm()
	form.AddField("Email")
	form.AddField("Password")

	form.mu.RLock()
	if len(form.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(form.Fields))
	}
	form.mu.RUnlock()
}

func TestFormValidation(t *testing.T) {
	form := NewForm()
	form.AddField("Email", func(s string) error {
		if len(s) < 5 {
			return NewTestError("too short")
		}
		return nil
	})

	form.updateFieldFocus()
	form.Fields[0].Input.Value = "test"
	form.validateAndSubmit()

	if form.Fields[0].Error == "" {
		t.Errorf("Expected validation error")
	}
}

func TestFormView(t *testing.T) {
	form := NewForm()
	form.AddField("Field1")

	view := form.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}
}

// TestDashboard tests dashboard
func TestDashboardBasic(t *testing.T) {
	dashboard := NewDashboard("Test Dashboard")

	if dashboard.Title != "Test Dashboard" {
		t.Errorf("Expected title 'Test Dashboard'")
	}
}

func TestDashboardAddComponent(t *testing.T) {
	dashboard := NewDashboard("Test")
	pb := NewProgressBar(100, "Progress")
	dashboard.AddComponent(pb)

	dashboard.mu.RLock()
	if len(dashboard.Components) != 1 {
		t.Errorf("Expected 1 component")
	}
	dashboard.mu.RUnlock()
}

func TestDashboardView(t *testing.T) {
	dashboard := NewDashboard("Test")
	pb := NewProgressBar(100, "Progress")
	pb.Update(75)
	dashboard.AddComponent(pb)

	view := dashboard.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}

	if !contains(view, "Test") {
		t.Errorf("Expected title in view")
	}
}

// TestInputHandler tests input handler
func TestInputHandlerBasic(t *testing.T) {
	handler := NewInputHandler()

	if handler.running {
		t.Errorf("Handler should not be running initially")
	}
}

func TestInputHandlerStop(t *testing.T) {
	handler := NewInputHandler()
	handler.mu.Lock()
	handler.running = true
	handler.mu.Unlock()

	handler.Stop()

	handler.mu.RLock()
	if handler.running {
		t.Errorf("Handler should be stopped")
	}
	handler.mu.RUnlock()
}

// Benchmark tests

func BenchmarkTextInputView(b *testing.B) {
	input := NewTextInput("Test")
	input.Value = "Benchmark Text"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = input.View()
	}
}

func BenchmarkListView(b *testing.B) {
	list := NewList()
	for i := 0; i < 20; i++ {
		list.AddItem("Item", nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = list.View()
	}
}

func BenchmarkTableView(b *testing.B) {
	table := NewTable([]string{"Col1", "Col2", "Col3"})
	for i := 0; i < 10; i++ {
		table.AddRow([]string{"A", "B", "C"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = table.View()
	}
}

func BenchmarkProgressBarView(b *testing.B) {
	pb := NewProgressBar(100, "Test")
	pb.Update(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pb.View()
	}
}

func BenchmarkFormView(b *testing.B) {
	form := NewForm()
	form.AddField("Field1")
	form.AddField("Field2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = form.View()
	}
}

// Helper functions

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type TestError struct {
	msg string
}

func (te *TestError) Error() string {
	return te.msg
}

func NewTestError(msg string) error {
	return &TestError{msg: msg}
}
