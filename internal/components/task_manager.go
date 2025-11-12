package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/internal/ui"
	"github.com/wyattlefevre/wydocli/logs"
)

type TaskManagerModel struct {
	tasks  []data.Task
	cursor int
}

type TaskUpdateMsg struct {
	Task data.Task
}

func (m *TaskManagerModel) WithTasks(tasks []data.Task) *TaskManagerModel {
	m.tasks = tasks
	return m
}

func (m *TaskManagerModel) Init() tea.Cmd {
	return nil
}

func (m *TaskManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ":
			logs.Logger.Println("space pressed")
			return m, func() tea.Msg {
				t := m.selectedTask()
				if t == nil {
					logs.Logger.Println("no selected task")
					return nil
				}
				// toggle status
				t.Done = !t.Done
				return TaskUpdateMsg{
					Task: *t,
				}
			}
		}
	}
	return m, nil
}

func (m *TaskManagerModel) View() string {
	var out string
	for i, task := range m.tasks {
		prefix := " "
		if i == m.cursor {
			prefix = "> "
		}
		out += prefix + ui.StyledTaskLine(task) + "\n"
	}
	return out
}

func (m *TaskManagerModel) selectedTask() *data.Task {
	if m.cursor >= 0 && m.cursor < len(m.tasks) {
		return &m.tasks[m.cursor]
	}
	return nil
}
