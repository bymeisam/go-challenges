package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Challenge 170: TUI Applications
// Interactive components, styling, navigation, real-time updates

// ===== 1. TUI Framework =====

type Model interface {
	Update(msg Message) Model
	View() string
}

type Message interface{}

type KeyMsg struct {
	Type KeyType
	Rune rune
}

type KeyType int

const (
	KeyUp KeyType = iota
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyTab
	KeyEscape
)

// ===== 2. Input Component =====

type TextInput struct {
	Value       string
	Placeholder string
	Focus       bool
	Cursor      int
	MaxLength   int
	Valid       bool
	mu          sync.RWMutex
}

func NewTextInput(placeholder string) *TextInput {
	return &TextInput{
		Placeholder: placeholder,
		MaxLength:   255,
		Valid:       true,
	}
}

func (ti *TextInput) Update(msg Message) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	switch m := msg.(type) {
	case KeyMsg:
		if !ti.Focus {
			return
		}

		switch m.Type {
		case KeyEnter:
			// Submit value
		default:
			if m.Type == KeyType(127) || m.Type == KeyType(8) { // Backspace
				if ti.Cursor > 0 {
					ti.Value = ti.Value[:ti.Cursor-1] + ti.Value[ti.Cursor:]
					ti.Cursor--
				}
			} else if m.Rune != 0 && len(ti.Value) < ti.MaxLength {
				ti.Value = ti.Value[:ti.Cursor] + string(m.Rune) + ti.Value[ti.Cursor:]
				ti.Cursor++
			}
		}
	}
}

func (ti *TextInput) View() string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	display := ti.Value
	if display == "" {
		display = fmt.Sprintf("[ %s ]", ti.Placeholder)
	} else {
		display = fmt.Sprintf("[ %s ]", display)
	}

	if ti.Focus {
		return fmt.Sprintf("\x1b[7m%s\x1b[0m", display)
	}
	return display
}

func (ti *TextInput) SetFocus(focus bool) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	ti.Focus = focus
}

// ===== 3. List Component =====

type ListItem struct {
	Label string
	Value interface{}
}

type List struct {
	Items       []*ListItem
	Cursor      int
	ViewPort    int
	Focus       bool
	Selected    []int
	MultiSelect bool
	mu          sync.RWMutex
}

func NewList() *List {
	return &List{
		Items:    make([]*ListItem, 0),
		ViewPort: 10,
	}
}

func (l *List) AddItem(label string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Items = append(l.Items, &ListItem{Label: label, Value: value})
}

func (l *List) Update(msg Message) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.Focus {
		return
	}

	switch m := msg.(type) {
	case KeyMsg:
		switch m.Type {
		case KeyUp:
			if l.Cursor > 0 {
				l.Cursor--
			}
		case KeyDown:
			if l.Cursor < len(l.Items)-1 {
				l.Cursor++
			}
		case KeyEnter:
			if l.MultiSelect {
				// Toggle selection
				found := false
				for i, s := range l.Selected {
					if s == l.Cursor {
						l.Selected = append(l.Selected[:i], l.Selected[i+1:]...)
						found = true
						break
					}
				}
				if !found {
					l.Selected = append(l.Selected, l.Cursor)
				}
			}
		}
	}
}

func (l *List) View() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var builder strings.Builder
	start := 0
	if l.Cursor > l.ViewPort {
		start = l.Cursor - l.ViewPort
	}

	end := start + l.ViewPort
	if end > len(l.Items) {
		end = len(l.Items)
	}

	for i := start; i < end; i++ {
		marker := "  "
		if i == l.Cursor {
			marker = "> "
		}

		if l.MultiSelect {
			selected := false
			for _, s := range l.Selected {
				if s == i {
					selected = true
					break
				}
			}
			if selected {
				marker += "[x] "
			} else {
				marker += "[ ] "
			}
		}

		builder.WriteString(marker)
		builder.WriteString(l.Items[i].Label)
		builder.WriteString("\n")
	}

	view := builder.String()
	if l.Focus {
		return fmt.Sprintf("\x1b[7m%s\x1b[0m", strings.TrimSpace(view))
	}
	return view
}

func (l *List) SetFocus(focus bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Focus = focus
}

func (l *List) GetSelectedValue() interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.Cursor >= 0 && l.Cursor < len(l.Items) {
		return l.Items[l.Cursor].Value
	}
	return nil
}

// ===== 4. Table Component =====

type TableRow struct {
	Cells []*TableCell
}

type TableCell struct {
	Value string
	Width int
}

type Table struct {
	Rows        []*TableRow
	Headers     []*TableCell
	CursorRow   int
	Focus       bool
	SortColumn  int
	SortDesc    bool
	mu          sync.RWMutex
}

func NewTable(headers []string) *Table {
	headerCells := make([]*TableCell, len(headers))
	maxWidth := 0
	for i, h := range headers {
		headerCells[i] = &TableCell{Value: h, Width: len(h)}
		if len(h) > maxWidth {
			maxWidth = len(h)
		}
	}

	return &Table{
		Headers: headerCells,
		Rows:    make([]*TableRow, 0),
	}
}

func (t *Table) AddRow(cells []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	row := &TableRow{Cells: make([]*TableCell, len(cells))}
	for i, cell := range cells {
		row.Cells[i] = &TableCell{Value: cell, Width: len(cell)}
	}
	t.Rows = append(t.Rows, row)
}

func (t *Table) View() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var builder strings.Builder

	// Header
	for _, header := range t.Headers {
		builder.WriteString(fmt.Sprintf("%-*s | ", header.Width, header.Value))
	}
	builder.WriteString("\n")

	// Separator
	for _, header := range t.Headers {
		builder.WriteString(strings.Repeat("-", header.Width))
		builder.WriteString("-+-")
	}
	builder.WriteString("\n")

	// Rows
	for i, row := range t.Rows {
		if i == t.CursorRow && t.Focus {
			builder.WriteString("\x1b[7m")
		}

		for j, cell := range row.Cells {
			if j < len(t.Headers) {
				builder.WriteString(fmt.Sprintf("%-*s | ", t.Headers[j].Width, cell.Value))
			}
		}

		if i == t.CursorRow && t.Focus {
			builder.WriteString("\x1b[0m")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func (t *Table) Update(msg Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.Focus {
		return
	}

	switch m := msg.(type) {
	case KeyMsg:
		switch m.Type {
		case KeyUp:
			if t.CursorRow > 0 {
				t.CursorRow--
			}
		case KeyDown:
			if t.CursorRow < len(t.Rows)-1 {
				t.CursorRow++
			}
		}
	}
}

// ===== 5. Progress Bar =====

type ProgressBar struct {
	Current   int64
	Total     int64
	Width     int
	Label     string
	StartTime time.Time
}

func NewProgressBar(total int64, label string) *ProgressBar {
	return &ProgressBar{
		Total:     total,
		Width:     50,
		Label:     label,
		StartTime: time.Now(),
	}
}

func (pb *ProgressBar) Update(current int64) {
	atomic.StoreInt64(&pb.Current, current)
}

func (pb *ProgressBar) View() string {
	current := atomic.LoadInt64(&pb.Current)
	total := atomic.LoadInt64(&pb.Total)

	if total == 0 {
		return ""
	}

	percentage := (current * 100) / total
	filledWidth := int((current * int64(pb.Width)) / total)

	bar := strings.Repeat("=", filledWidth) + strings.Repeat("-", pb.Width-filledWidth)

	elapsed := time.Since(pb.StartTime)
	var eta time.Duration
	if current > 0 {
		eta = time.Duration((elapsed.Nanoseconds() / int64(current)) * (total - current))
	}

	return fmt.Sprintf("%s [%s] %d%% (%d/%d) ETA: %v",
		pb.Label, bar, percentage, current, total, eta)
}

// ===== 6. Form Component =====

type FormField struct {
	Label      string
	Input      *TextInput
	Validators []func(string) error
	Error      string
}

type Form struct {
	Fields      []*FormField
	CurrentField int
	Values      map[string]string
	Submitted   bool
	mu          sync.RWMutex
}

func NewForm() *Form {
	return &Form{
		Fields: make([]*FormField, 0),
		Values: make(map[string]string),
	}
}

func (f *Form) AddField(label string, validators ...func(string) error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	field := &FormField{
		Label:      label,
		Input:      NewTextInput(label),
		Validators: validators,
	}

	f.Fields = append(f.Fields, field)
}

func (f *Form) Update(msg Message) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.CurrentField >= len(f.Fields) {
		return
	}

	switch m := msg.(type) {
	case KeyMsg:
		switch m.Type {
		case KeyTab:
			f.CurrentField = (f.CurrentField + 1) % len(f.Fields)
			f.updateFieldFocus()
		case KeyEnter:
			if f.CurrentField == len(f.Fields)-1 {
				f.validateAndSubmit()
			} else {
				f.CurrentField++
				f.updateFieldFocus()
			}
		default:
			f.Fields[f.CurrentField].Input.Update(msg)
		}
	}
}

func (f *Form) updateFieldFocus() {
	for i, field := range f.Fields {
		field.Input.SetFocus(i == f.CurrentField)
	}
}

func (f *Form) validateAndSubmit() {
	for _, field := range f.Fields {
		field.Input.mu.RLock()
		value := field.Input.Value
		field.Input.mu.RUnlock()

		for _, validator := range field.Validators {
			if err := validator(value); err != nil {
				field.Error = err.Error()
				return
			}
		}

		f.Values[field.Label] = value
	}

	f.Submitted = true
}

func (f *Form) View() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var builder strings.Builder

	for i, field := range f.Fields {
		builder.WriteString(fmt.Sprintf("%s: ", field.Label))
		builder.WriteString(field.Input.View())
		if field.Error != "" {
			builder.WriteString(fmt.Sprintf(" \x1b[31m%s\x1b[0m", field.Error))
		}
		builder.WriteString("\n")

		if i == f.CurrentField {
			builder.WriteString("> ")
		}
	}

	return builder.String()
}

// ===== 7. Dashboard =====

type Dashboard struct {
	Title      string
	Components []interface{}
	mu         sync.RWMutex
}

func NewDashboard(title string) *Dashboard {
	return &Dashboard{
		Title:      title,
		Components: make([]interface{}, 0),
	}
}

func (d *Dashboard) AddComponent(component interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Components = append(d.Components, component)
}

func (d *Dashboard) View() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n=== %s ===\n\n", d.Title))

	for _, component := range d.Components {
		switch c := component.(type) {
		case *ProgressBar:
			builder.WriteString(c.View())
		case *List:
			builder.WriteString(c.View())
		case *Table:
			builder.WriteString(c.View())
		case *Form:
			builder.WriteString(c.View())
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// ===== 8. Input Handler =====

type InputHandler struct {
	mu      sync.RWMutex
	running bool
	stopChan chan struct{}
}

func NewInputHandler() *InputHandler {
	return &InputHandler{
		stopChan: make(chan struct{}),
	}
}

func (ih *InputHandler) Start(callback func(Message)) {
	ih.mu.Lock()
	ih.running = true
	ih.mu.Unlock()

	// Simulate keyboard input handling
	// In real implementation, would use termios/tcgetattr
}

func (ih *InputHandler) Stop() {
	ih.mu.Lock()
	defer ih.mu.Unlock()

	if ih.running {
		close(ih.stopChan)
		ih.running = false
	}
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== TUI Applications ===\n")

	// 1. TextInput Component
	fmt.Println("1. Text Input Component")
	input := NewTextInput("Enter your name")
	input.SetFocus(true)
	input.Value = "Alice"
	fmt.Printf("Input: %s\n\n", input.View())

	// 2. List Component
	fmt.Println("2. List Component")
	list := NewList()
	list.AddItem("Option 1", "value1")
	list.AddItem("Option 2", "value2")
	list.AddItem("Option 3", "value3")
	list.SetFocus(true)
	fmt.Printf("List:\n%s\n\n", list.View())

	// 3. Table Component
	fmt.Println("3. Table Component")
	table := NewTable([]string{"Name", "Age", "City"})
	table.AddRow([]string{"Alice", "30", "NYC"})
	table.AddRow([]string{"Bob", "25", "LA"})
	table.AddRow([]string{"Charlie", "35", "Chicago"})
	table.SetFocus(true)
	fmt.Printf("Table:\n%s\n\n", table.View())

	// 4. Progress Bar
	fmt.Println("4. Progress Bar")
	pb := NewProgressBar(100, "Download")
	for i := 0; i <= 100; i += 20 {
		pb.Update(int64(i))
		fmt.Printf("\r%s", pb.View())
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("\n\n")

	// 5. Form
	fmt.Println("5. Form Component")
	form := NewForm()
	form.AddField("Email", func(s string) error {
		if len(s) < 5 {
			return fmt.Errorf("email too short")
		}
		return nil
	})
	form.AddField("Password", func(s string) error {
		if len(s) < 8 {
			return fmt.Errorf("password must be 8+ chars")
		}
		return nil
	})
	form.updateFieldFocus()
	fmt.Printf("Form:\n%s\n\n", form.View())

	// 6. Dashboard
	fmt.Println("6. Dashboard")
	dashboard := NewDashboard("System Monitor")

	pb2 := NewProgressBar(100, "CPU Usage")
	pb2.Update(65)
	dashboard.AddComponent(pb2)

	pb3 := NewProgressBar(100, "Memory Usage")
	pb3.Update(45)
	dashboard.AddComponent(pb3)

	fmt.Println(dashboard.View())

	// 7. Features Summary
	fmt.Println("7. TUI Features")
	fmt.Println("  - Text input with cursor")
	fmt.Println("  - Selectable lists")
	fmt.Println("  - Data tables with focus")
	fmt.Println("  - Progress bars with ETA")
	fmt.Println("  - Forms with validation")
	fmt.Println("  - Dashboard layout")
	fmt.Println("  - ANSI color support")
	fmt.Println("  - Keyboard navigation")

	fmt.Println("\n=== Complete ===")
}
