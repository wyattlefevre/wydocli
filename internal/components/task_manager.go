package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/internal/ui"
	"github.com/wyattlefevre/wydocli/logs"
)

type InputMode int

const (
	NormalInputMode InputMode = iota
	FilterInputMode
	GroupInputMode
	SortInputMode
	SearchInputMode
)

type GroupMode int

const (
	GroupNone GroupMode = iota
	GroupProject
	GroupDate
	GroupPriority
)

type SortMode int

const (
	SortNone SortMode = iota
	SortProject
	SortDate
	SortPriority
)

type TaskManagerModel struct {
	tasks        []data.Task
	displayTasks []data.Task
	cursor       int
	inputMode    InputMode

	searchStr     string
	projectFilter string
	dateFilter    string
	groupMode     GroupMode
	sortMode      SortMode
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
		if msg.String() == "esc" {
			m.inputMode = NormalInputMode
			return m, nil
		}
		switch m.inputMode {
		case NormalInputMode:
			return m.handleNormalInput(msg.String())
		case FilterInputMode:
			return m.handleFilterInput(msg.String())
		}
	}
	return m, nil
}

func (m *TaskManagerModel) View() string {
	var out string
	out += modeLine(m.inputMode)
	for i, task := range m.tasks {
		prefix := " "
		if i == m.cursor {
			prefix = "> "
		}
		out += prefix + ui.StyledTaskLine(task) + "\n"
	}
	return out
}

func modeLine(mode InputMode) string {
	out := ""
	switch mode {
	case NormalInputMode:
		return out + "[Normal]\n"
	case FilterInputMode:
		return out + "[Filter]\n"
	case GroupInputMode:
		return out + "[Group]\n"
	case SortInputMode:
		return out + "[Sort]\n"
	case SearchInputMode:
		return out + "[Search]\n"
	}
	return "MODE ERR"
}

func (m *TaskManagerModel) selectedTask() *data.Task {
	if m.cursor >= 0 && m.cursor < len(m.tasks) {
		return &m.tasks[m.cursor]
	}
	return nil
}

func (m *TaskManagerModel) handleFilterInput(key string) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *TaskManagerModel) handleNormalInput(key string) (tea.Model, tea.Cmd) {
	switch key {
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
	case "f":
		m.inputMode = FilterInputMode
	case "g":
		m.inputMode = GroupInputMode
	case "s":
		m.inputMode = SortInputMode
	case "/":
		m.inputMode = SearchInputMode
	}
	return m, nil
}
