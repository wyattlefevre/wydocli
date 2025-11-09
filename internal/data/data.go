package data

import (
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Messages emitted to the app when data changes
type TasksUpdatedMsg struct {
	Tasks []Task
}

// internal shared state
var (
	mu       sync.RWMutex
	allTasks []Task
	todoPath = "./todo.txt" // or configurable
)

// Load all tasks from the todo.txt file
func LoadTasks() ([]Task, error) {
	mu.Lock()
	defer mu.Unlock()

	bytes, err := os.ReadFile(todoPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")
	var tasks []Task
	for _, line := range lines {
		if line == "" {
			continue
		}
		tasks = append(tasks, parseTask(line))
	}

	allTasks = tasks
	return tasks, nil
}

// Get a copy of the in-memory task list
func GetTasks() []Task {
	mu.RLock()
	defer mu.RUnlock()

	tasks := make([]Task, len(allTasks))
	copy(tasks, allTasks)
	return tasks
}

// Add a task and write to disk, then emit a refresh message
func AddTaskCmd(description string) tea.Cmd {
	return func() tea.Msg {
		mu.Lock()
		defer mu.Unlock()

		// append to file
		f, err := os.ReadFile(todoPath)
		if err != nil {
			return err
		}

		updated := strings.TrimSpace(string(f)) + "\n" + description + "\n"
		if err := os.WriteFile(todoPath, []byte(updated), 0644); err != nil {
			return err
		}

		// refresh tasks in memory
		tasks, err := LoadTasks()
		if err != nil {
			return err
		}

		// emit a message to trigger UI update
		return TasksUpdatedMsg{Tasks: tasks}
	}
}

// Optionally poll for changes every few seconds
func PollTasksCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		tasks, _ := LoadTasks()
		return TasksUpdatedMsg{Tasks: tasks}
	})
}

// simple todo.txt parser stub
func parseTask(line string) Task {
	done := strings.HasPrefix(line, "x ")
	return Task{
		Name: line,
		Done:        done,
	}
}
