package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
)

type ProjectManagerModel struct {
	projects map[string]data.Project
}

func (m *ProjectManagerModel) Init() tea.Cmd {
	return nil
}

func (m *ProjectManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Implement update logic here
	return m, nil
}

func (m *ProjectManagerModel) View() string {
	// Implement view logic here
	return "Project Manager"
}
