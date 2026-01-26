# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Run the application
go run ./cmd/wydocli

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/data

# Run a specific test
go test ./internal/data -run TestParseTask_TableDriven
```

## Architecture

wydoCLI is a terminal-based task manager built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI framework) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling). It uses the todo.txt format for task storage.

### Package Structure

- **cmd/wydocli/main.go** - Entry point, initializes the Bubble Tea program
- **internal/app/** - Main application model (`AppModel`) that coordinates views and handles data loading/saving
- **internal/components/** - Bubble Tea models for different views:
  - `TaskManagerModel` - Primary task list view with vim-style navigation (j/k), filtering, sorting
  - `ProjectManagerModel` - Project list view (stub)
  - `TaskEditorModel` - Task editing (stub)
- **internal/data/** - Data layer for parsing and persisting tasks/projects
- **internal/ui/** - Styling utilities for rendering task lines
- **logs/** - Global logger that writes to `debug.log`

### Data Flow

1. `AppModel.Init()` loads tasks from todo.txt/done.txt files and projects from a project directory
2. User interactions produce `tea.Msg` types (e.g., `TaskUpdateMsg`)
3. `AppModel.Update()` handles messages, writes changes to disk, and reloads data
4. Views are switched via global keys: `P` (Projects), `T` (Tasks), `F` (Files - not implemented)

### Task Parsing

Tasks follow the [todo.txt format](http://todotxt.org/). The parser in `internal/data/task.go`:
- Supports priority (A-F), completion dates, creation dates
- Extracts projects (`+project`), contexts (`@context`), and key:value tags
- Uses `Task.String()` to serialize back to todo.txt format
- Validates round-trip parsing (parsed task must serialize back to original)

### Configuration via Environment Variables

- `TODO_DIR` - Base directory for todo files (default: `$HOME`)
- `TODO_FILE` - Path to todo.txt (default: `$TODO_DIR/todo.txt`)
- `DONE_FILE` - Path to done.txt (default: `$TODO_DIR/done.txt`)
- `TODO_PROJ_DIR` - Directory containing project note files (default: `$TODO_DIR/todo_projects`)
