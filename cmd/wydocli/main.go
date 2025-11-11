package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/app"
)

func main() {
	app := &app.AppModel{}
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
