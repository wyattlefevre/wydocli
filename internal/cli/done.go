package cli

import (
	"fmt"
	"os"

	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/internal/service"
)

func runDone(args []string, svc service.TaskService) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: task ID required")
		fmt.Fprintln(os.Stderr, "Usage: wydo done <task-id>")
		return 1
	}

	taskID := args[0]

	// Try to find the task first (supports partial ID matching)
	task, err := findTaskByPartialID(svc, taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if task.Done {
		fmt.Printf("Task already completed: %s\n", task.Name)
		return 0
	}

	err = svc.Complete(task.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
		return 1
	}

	fmt.Printf("Completed: %s\n", task.Name)
	return 0
}

// findTaskByPartialID finds a task by full or partial ID
func findTaskByPartialID(svc service.TaskService, partialID string) (*data.Task, error) {
	tasks, err := svc.List()
	if err != nil {
		return nil, err
	}

	var matches []data.Task
	for _, t := range tasks {
		if t.ID == partialID || (len(partialID) >= 4 && len(t.ID) >= len(partialID) && t.ID[:len(partialID)] == partialID) {
			matches = append(matches, t)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no task found with ID: %s", partialID)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple tasks match ID '%s', please be more specific", partialID)
	}

	return &matches[0], nil
}
