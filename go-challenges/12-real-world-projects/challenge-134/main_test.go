package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewTodoApp(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")

	app, err := NewTodoApp(tmpFile)
	if err != nil {
		t.Fatalf("NewTodoApp failed: %v", err)
	}

	if app == nil {
		t.Fatal("Expected app to be non-nil")
	}

	if len(app.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(app.Tasks))
	}
}

func TestAddTask(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	tests := []struct {
		name     string
		text     string
		priority Priority
		wantErr  bool
	}{
		{"Valid task", "Buy groceries", PriorityMedium, false},
		{"High priority", "Urgent task", PriorityHigh, false},
		{"Low priority", "Maybe later", PriorityLow, false},
		{"Empty text", "", PriorityMedium, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := app.Add(tt.text, tt.priority)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Add failed: %v", err)
			}

			if task.Text != tt.text {
				t.Errorf("Expected text %q, got %q", tt.text, task.Text)
			}

			if task.Priority != tt.priority {
				t.Errorf("Expected priority %v, got %v", tt.priority, task.Priority)
			}

			if task.Completed {
				t.Error("New task should not be completed")
			}

			if task.ID <= 0 {
				t.Error("Task ID should be positive")
			}
		})
	}
}

func TestCompleteTask(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	task, _ := app.Add("Test task", PriorityMedium)

	// Complete the task
	err := app.Complete(task.ID)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Verify it's completed
	if !app.Tasks[0].Completed {
		t.Error("Task should be marked as completed")
	}

	// Try to complete again
	err = app.Complete(task.ID)
	if err == nil {
		t.Error("Expected error when completing already completed task")
	}

	// Try to complete non-existent task
	err = app.Complete(999)
	if err == nil {
		t.Error("Expected error when completing non-existent task")
	}
}

func TestDeleteTask(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	task1, _ := app.Add("Task 1", PriorityMedium)
	task2, _ := app.Add("Task 2", PriorityHigh)

	// Delete first task
	err := app.Delete(task1.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if len(app.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(app.Tasks))
	}

	if app.Tasks[0].ID != task2.ID {
		t.Error("Wrong task was deleted")
	}

	// Try to delete non-existent task
	err = app.Delete(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent task")
	}
}

func TestListTasks(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	task1, _ := app.Add("Task 1", PriorityMedium)
	task2, _ := app.Add("Task 2", PriorityHigh)
	app.Complete(task1.ID)

	// List all tasks
	allTasks := app.List(true)
	if len(allTasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(allTasks))
	}

	// List only pending tasks
	pendingTasks := app.List(false)
	if len(pendingTasks) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
	}

	if pendingTasks[0].ID != task2.ID {
		t.Error("Wrong pending task returned")
	}
}

func TestClearTasks(t *testing.T) {
	t.Run("Clear all tasks", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
		app, _ := NewTodoApp(tmpFile)

		app.Add("Task 1", PriorityMedium)
		app.Add("Task 2", PriorityHigh)
		app.Add("Task 3", PriorityLow)

		count, err := app.Clear(false)
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}

		if count != 3 {
			t.Errorf("Expected 3 cleared tasks, got %d", count)
		}

		if len(app.Tasks) != 0 {
			t.Errorf("Expected 0 tasks, got %d", len(app.Tasks))
		}
	})

	t.Run("Clear only completed tasks", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
		app, _ := NewTodoApp(tmpFile)

		task1, _ := app.Add("Task 1", PriorityMedium)
		app.Add("Task 2", PriorityHigh)
		task3, _ := app.Add("Task 3", PriorityLow)

		app.Complete(task1.ID)
		app.Complete(task3.ID)

		count, err := app.Clear(true)
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 cleared tasks, got %d", count)
		}

		if len(app.Tasks) != 1 {
			t.Errorf("Expected 1 remaining task, got %d", len(app.Tasks))
		}
	})
}

func TestPersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")

	// Create app and add tasks
	app1, _ := NewTodoApp(tmpFile)
	task1, _ := app1.Add("Persistent task", PriorityHigh)
	app1.Complete(task1.ID)

	// Create new app instance from same file
	app2, err := NewTodoApp(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load persisted data: %v", err)
	}

	if len(app2.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(app2.Tasks))
	}

	task := app2.Tasks[0]
	if task.Text != "Persistent task" {
		t.Errorf("Expected text %q, got %q", "Persistent task", task.Text)
	}

	if !task.Completed {
		t.Error("Task should be completed")
	}

	if task.Priority != PriorityHigh {
		t.Errorf("Expected priority %v, got %v", PriorityHigh, task.Priority)
	}
}

func TestTaskIDIncrement(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	task1, _ := app.Add("Task 1", PriorityMedium)
	task2, _ := app.Add("Task 2", PriorityMedium)
	task3, _ := app.Add("Task 3", PriorityMedium)

	if task1.ID != 1 {
		t.Errorf("Expected ID 1, got %d", task1.ID)
	}

	if task2.ID != 2 {
		t.Errorf("Expected ID 2, got %d", task2.ID)
	}

	if task3.ID != 3 {
		t.Errorf("Expected ID 3, got %d", task3.ID)
	}

	// Delete middle task
	app.Delete(task2.ID)

	// Next task should still get ID 4
	task4, _ := app.Add("Task 4", PriorityMedium)
	if task4.ID != 4 {
		t.Errorf("Expected ID 4, got %d", task4.ID)
	}
}

func TestTaskTimestamp(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	before := time.Now()
	task, _ := app.Add("Timed task", PriorityMedium)
	after := time.Now()

	if task.CreatedAt.Before(before) || task.CreatedAt.After(after) {
		t.Error("Task timestamp is out of expected range")
	}
}

func TestEmptyList(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	tasks := app.List(true)
	if len(tasks) != 0 {
		t.Errorf("Expected empty list, got %d tasks", len(tasks))
	}

	tasks = app.List(false)
	if len(tasks) != 0 {
		t.Errorf("Expected empty list, got %d tasks", len(tasks))
	}
}

func TestFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new_tasks.json")

	// File should not exist
	if _, err := os.Stat(tmpFile); err == nil {
		t.Fatal("File should not exist before creating app")
	}

	// Create app
	_, err := NewTodoApp(tmpFile)
	if err != nil {
		t.Fatalf("NewTodoApp failed: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("File should exist after creating app")
	}
}

func TestMultiplePriorities(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_tasks.json")
	app, _ := NewTodoApp(tmpFile)

	taskHigh, _ := app.Add("High priority", PriorityHigh)
	taskMed, _ := app.Add("Medium priority", PriorityMedium)
	taskLow, _ := app.Add("Low priority", PriorityLow)

	if taskHigh.Priority != PriorityHigh {
		t.Error("High priority task has wrong priority")
	}

	if taskMed.Priority != PriorityMedium {
		t.Error("Medium priority task has wrong priority")
	}

	if taskLow.Priority != PriorityLow {
		t.Error("Low priority task has wrong priority")
	}
}
