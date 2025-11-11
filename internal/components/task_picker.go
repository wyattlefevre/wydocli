package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
)

type TaskPickerModel struct {
	tasks []data.Task
}

func NewTaskPickerModel(tasks []data.Task) *TaskPickerModel {
	return &TaskPickerModel{
		tasks: tasks,
	}
}

func (m *TaskPickerModel) Init() tea.Cmd {
	return nil
}

func (m *TaskPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TaskPickerModel) View() string {
	var out string
	for _, task := range m.tasks {
		out += task.String() + "\n"
	}
	return out
}
