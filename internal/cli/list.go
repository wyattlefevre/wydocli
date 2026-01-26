package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/internal/service"
)

func runList(args []string, svc service.TaskService) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	project := fs.String("p", "", "Filter by project")
	context := fs.String("c", "", "Filter by context")
	showDone := fs.Bool("done", false, "Show only completed tasks")
	showAll := fs.Bool("all", false, "Show all tasks including completed")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	var tasks []data.Task
	var err error

	// Get base task list
	if *showDone {
		tasks, err = svc.ListDone()
	} else if *showAll {
		tasks, err = svc.List()
	} else {
		tasks, err = svc.ListPending()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		return 1
	}

	// Apply filters
	if *project != "" {
		tasks = filterByProject(tasks, *project)
	}
	if *context != "" {
		tasks = filterByContext(tasks, *context)
	}

	// Print tasks
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return 0
	}

	for _, t := range tasks {
		printTask(t)
	}

	fmt.Printf("\n%d task(s)\n", len(tasks))
	return 0
}

func filterByProject(tasks []data.Task, project string) []data.Task {
	var filtered []data.Task
	for _, t := range tasks {
		if t.HasProject(project) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func filterByContext(tasks []data.Task, context string) []data.Task {
	var filtered []data.Task
	for _, t := range tasks {
		if t.HasContext(context) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func printTask(t data.Task) {
	// Format: [ID] (Priority) Task description +project @context
	status := " "
	if t.Done {
		status = "x"
	}

	priority := ""
	if t.Priority != 0 {
		priority = fmt.Sprintf("(%c) ", t.Priority)
	}

	fmt.Printf("[%s] %s %s%s\n", t.ID[:7], status, priority, t.Name)

	// Print projects and contexts on same line if present
	var meta []string
	for _, p := range t.Projects {
		meta = append(meta, "+"+p)
	}
	for _, c := range t.Contexts {
		meta = append(meta, "@"+c)
	}
	if len(meta) > 0 {
		fmt.Printf("        ")
		for _, m := range meta {
			fmt.Printf("%s ", m)
		}
		fmt.Println()
	}
}
