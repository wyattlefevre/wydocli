package cli

import (
	"fmt"
	"os"

	"github.com/wyattlefevre/wydocli/internal/service"
)

func runDelete(args []string, svc service.TaskService) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: task ID required")
		fmt.Fprintln(os.Stderr, "Usage: wydo delete <task-id>")
		return 1
	}

	taskID := args[0]

	// Try to find the task first (supports partial ID matching)
	task, err := findTaskByPartialID(svc, taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	err = svc.Delete(task.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		return 1
	}

	fmt.Printf("Deleted: %s\n", task.Name)
	return 0
}
