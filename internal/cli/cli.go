package cli

import (
	"fmt"
	"os"

	"github.com/wyattlefevre/wydocli/internal/service"
)

// Run executes the CLI with the given arguments.
// Returns an exit code (0 for success, non-zero for errors).
func Run(args []string, svc service.TaskService) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "add", "a":
		return runAdd(cmdArgs, svc)
	case "list", "ls", "l":
		return runList(cmdArgs, svc)
	case "done", "do", "d":
		return runDone(cmdArgs, svc)
	case "delete", "rm", "del":
		return runDelete(cmdArgs, svc)
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Println(`wydo - A command-line task manager using todo.txt format

Usage: wydo [command] [arguments]

Commands:
  add, a      Add a new task
              wydo add "Task description +project @context"

  list, ls, l List tasks
              wydo list              # List all pending tasks
              wydo list --all        # List all tasks including done
              wydo list -p project   # Filter by project
              wydo list -c context   # Filter by context
              wydo list --done       # List only completed tasks

  done, do, d Mark a task as complete
              wydo done <task-id>

  delete, rm  Delete a task
              wydo delete <task-id>

  help        Show this help message

Running wydo without arguments launches the interactive TUI.`)
}
