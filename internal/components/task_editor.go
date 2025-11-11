package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

type TaskEditorModel struct {
	
}

func (a *TaskEditorModel) Init() tea.Cmd {
	return nil
}

func (a *TaskEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a, nil
}

func (a *TaskEditorModel) View() string {
	return ""
}
