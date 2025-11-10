# Challenge 134: CLI Todo App

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Build a complete command-line todo application using the Cobra library. This project demonstrates CLI application development, persistent storage, command parsing, and user interaction patterns.

## Features

- **Multiple Commands**: add, list, complete, delete, clear
- **Persistent Storage**: JSON file-based storage
- **Task Management**: ID-based task identification
- **Status Tracking**: Mark tasks as complete/incomplete
- **Filtering**: View all tasks or only pending ones
- **Color Output**: Use colors for better UX
- **Timestamps**: Track when tasks were created
- **Priority Levels**: Support high, medium, low priority
- **Error Handling**: Graceful error messages

## Commands

```bash
todo add "Buy groceries" --priority high
todo list [--all]
todo complete <id>
todo delete <id>
todo clear [--completed]
```

## Requirements

1. Implement using Cobra for command handling
2. Store tasks in JSON format
3. Auto-assign sequential IDs
4. Display formatted output
5. Handle edge cases (invalid IDs, empty lists, etc.)

## Example Usage

```bash
$ todo add "Write documentation"
✓ Task added: #1 "Write documentation"

$ todo add "Review PR" --priority high
✓ Task added: #2 "Review PR" (High Priority)

$ todo list
ID  Status  Priority  Task
1   [ ]     Medium    Write documentation
2   [ ]     High      Review PR

$ todo complete 1
✓ Task #1 marked as complete

$ todo list --all
ID  Status  Priority  Task
1   [✓]     Medium    Write documentation
2   [ ]     High      Review PR
```

## Learning Objectives

- CLI application architecture
- Command pattern implementation
- File-based persistence
- User input validation
- Error handling in CLI apps
- Cobra library usage
- JSON serialization
