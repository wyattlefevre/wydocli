package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/app"
	"github.com/wyattlefevre/wydocli/logs"
)

func main() {
	app := &app.AppModel{}
	logs.Logger.Println("Starting app")
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
