package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Priority represents task priority levels
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Task represents a todo item
type Task struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	Priority  Priority  `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// TodoApp manages the todo list
type TodoApp struct {
	Tasks    []Task `json:"tasks"`
	filepath string
	nextID   int
}

// NewTodoApp creates a new todo application
func NewTodoApp(filepath string) (*TodoApp, error) {
	app := &TodoApp{
		Tasks:    make([]Task, 0),
		filepath: filepath,
		nextID:   1,
	}

	if err := app.load(); err != nil {
		return nil, err
	}

	return app, nil
}

// load reads tasks from the JSON file
func (app *TodoApp) load() error {
	// Create file if it doesn't exist
	if _, err := os.Stat(app.filepath); os.IsNotExist(err) {
		return app.save()
	}

	data, err := os.ReadFile(app.filepath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &app.Tasks); err != nil {
		return err
	}

	// Update nextID
	for _, task := range app.Tasks {
		if task.ID >= app.nextID {
			app.nextID = task.ID + 1
		}
	}

	return nil
}

// save writes tasks to the JSON file
func (app *TodoApp) save() error {
	data, err := json.MarshalIndent(app.Tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(app.filepath, data, 0644)
}

// Add adds a new task
func (app *TodoApp) Add(text string, priority Priority) (*Task, error) {
	if text == "" {
		return nil, fmt.Errorf("task text cannot be empty")
	}

	task := Task{
		ID:        app.nextID,
		Text:      text,
		Completed: false,
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	app.Tasks = append(app.Tasks, task)
	app.nextID++

	if err := app.save(); err != nil {
		return nil, err
	}

	return &task, nil
}

// List returns all tasks or only pending ones
func (app *TodoApp) List(showAll bool) []Task {
	if showAll {
		return app.Tasks
	}

	pending := make([]Task, 0)
	for _, task := range app.Tasks {
		if !task.Completed {
			pending = append(pending, task)
		}
	}
	return pending
}

// Complete marks a task as complete
func (app *TodoApp) Complete(id int) error {
	for i := range app.Tasks {
		if app.Tasks[i].ID == id {
			if app.Tasks[i].Completed {
				return fmt.Errorf("task #%d is already completed", id)
			}
			app.Tasks[i].Completed = true
			return app.save()
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

// Delete removes a task
func (app *TodoApp) Delete(id int) error {
	for i := range app.Tasks {
		if app.Tasks[i].ID == id {
			app.Tasks = append(app.Tasks[:i], app.Tasks[i+1:]...)
			return app.save()
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

// Clear removes completed tasks or all tasks
func (app *TodoApp) Clear(completedOnly bool) (int, error) {
	if !completedOnly {
		count := len(app.Tasks)
		app.Tasks = make([]Task, 0)
		if err := app.save(); err != nil {
			return 0, err
		}
		return count, nil
	}

	remaining := make([]Task, 0)
	count := 0
	for _, task := range app.Tasks {
		if task.Completed {
			count++
		} else {
			remaining = append(remaining, task)
		}
	}

	app.Tasks = remaining
	if err := app.save(); err != nil {
		return 0, err
	}

	return count, nil
}

// CLI colors
var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	cyan   = color.New(color.FgCyan)
	bold   = color.New(color.Bold)
)

func main() {
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".todo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	todoFile := filepath.Join(configDir, "tasks.json")

	// Create app
	app, err := NewTodoApp(todoFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Root command
	var rootCmd = &cobra.Command{
		Use:   "todo",
		Short: "A simple CLI todo application",
		Long:  "A command-line todo application with persistent storage",
	}

	// Add command
	var priority string
	var addCmd = &cobra.Command{
		Use:   "add [task]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			p := Priority(priority)
			if p != PriorityHigh && p != PriorityMedium && p != PriorityLow {
				p = PriorityMedium
			}

			task, err := app.Add(args[0], p)
			if err != nil {
				red.Fprintf(os.Stderr, "✗ Error: %v\n", err)
				os.Exit(1)
			}

			green.Printf("✓ Task added: #%d \"%s\"", task.ID, task.Text)
			if task.Priority != PriorityMedium {
				fmt.Printf(" (%s Priority)", string(task.Priority))
			}
			fmt.Println()
		},
	}
	addCmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Task priority (high, medium, low)")

	// List command
	var showAll bool
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasks := app.List(showAll)
			if len(tasks) == 0 {
				yellow.Println("No tasks found")
				return
			}

			// Header
			bold.Printf("%-4s %-8s %-10s %s\n", "ID", "Status", "Priority", "Task")
			fmt.Println("------------------------------------------------------------")

			// Tasks
			for _, task := range tasks {
				status := "[ ]"
				if task.Completed {
					status = "[✓]"
				}

				priorityStr := string(task.Priority)
				switch task.Priority {
				case PriorityHigh:
					red.Printf("%-4d %-8s %-10s %s\n", task.ID, status, priorityStr, task.Text)
				case PriorityLow:
					cyan.Printf("%-4d %-8s %-10s %s\n", task.ID, status, priorityStr, task.Text)
				default:
					fmt.Printf("%-4d %-8s %-10s %s\n", task.ID, status, priorityStr, task.Text)
				}
			}
		},
	}
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all tasks including completed")

	// Complete command
	var completeCmd = &cobra.Command{
		Use:   "complete [id]",
		Short: "Mark a task as complete",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				red.Fprintf(os.Stderr, "✗ Error: invalid task ID\n")
				os.Exit(1)
			}

			if err := app.Complete(id); err != nil {
				red.Fprintf(os.Stderr, "✗ Error: %v\n", err)
				os.Exit(1)
			}

			green.Printf("✓ Task #%d marked as complete\n", id)
		},
	}

	// Delete command
	var deleteCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				red.Fprintf(os.Stderr, "✗ Error: invalid task ID\n")
				os.Exit(1)
			}

			if err := app.Delete(id); err != nil {
				red.Fprintf(os.Stderr, "✗ Error: %v\n", err)
				os.Exit(1)
			}

			green.Printf("✓ Task #%d deleted\n", id)
		},
	}

	// Clear command
	var completedOnly bool
	var clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear tasks",
		Run: func(cmd *cobra.Command, args []string) {
			count, err := app.Clear(completedOnly)
			if err != nil {
				red.Fprintf(os.Stderr, "✗ Error: %v\n", err)
				os.Exit(1)
			}

			if completedOnly {
				green.Printf("✓ Cleared %d completed task(s)\n", count)
			} else {
				green.Printf("✓ Cleared all %d task(s)\n", count)
			}
		},
	}
	clearCmd.Flags().BoolVarP(&completedOnly, "completed", "c", false, "Clear only completed tasks")

	rootCmd.AddCommand(addCmd, listCmd, completeCmd, deleteCmd, clearCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
