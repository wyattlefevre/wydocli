package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/app"
	"github.com/wyattlefevre/wydocli/internal/cli"
	"github.com/wyattlefevre/wydocli/internal/config"
	"github.com/wyattlefevre/wydocli/internal/service"
	"github.com/wyattlefevre/wydocli/logs"
)

func main() {
	// Define global flags
	todoDir := flag.String("d", "", "Path to todo directory (overrides config file and env vars)")
	flag.StringVar(todoDir, "todo-dir", "", "Path to todo directory (overrides config file and env vars)")

	// Parse flags, but stop at first non-flag argument (the subcommand)
	flag.Parse()

	// Set CLI flags before loading config
	if *todoDir != "" {
		config.SetCLIFlags(config.CLIFlags{TodoDir: *todoDir})
	}

	// Load configuration
	_, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize the task service
	svc, err := service.NewTaskService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing service: %v\n", err)
		os.Exit(1)
	}

	// Remaining args after flag parsing
	args := flag.Args()

	if len(args) > 0 {
		// CLI mode
		exitCode := cli.Run(args, svc)
		os.Exit(exitCode)
	}

	// TUI mode
	logs.Logger.Println("Starting app in TUI mode")
	appModel := app.NewAppModelWithService(svc)
	p := tea.NewProgram(appModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
