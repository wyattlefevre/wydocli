package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/wyattlefevre/wydocli/internal/service"
)

func runAdd(args []string, svc service.TaskService) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: task description required")
		fmt.Fprintln(os.Stderr, "Usage: wydo add \"Task description +project @context\"")
		return 1
	}

	// Join all arguments as the task line (allows for unquoted input)
	rawLine := strings.Join(args, " ")

	task, err := svc.Add(rawLine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding task: %v\n", err)
		return 1
	}

	fmt.Printf("Added: %s\n", task.String())
	fmt.Printf("ID: %s\n", task.ID)
	return 0
}
