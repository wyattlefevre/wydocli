package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/components"
	"github.com/wyattlefevre/wydocli/internal/data"
)

type AppModel struct {
	stack         []tea.Model // TODO: might not need a stack. would make it easier to explicitly update and control components. in fact, it would be nice for preserving state. Only need one task picker loaded
	todoFileTasks []data.Task
	doneFileTasks []data.Task
	projectMap    map[string]data.Project
}

type ParseTaskMismatchMsg struct {
	Err *data.ParseTaskMismatchError
}

type initDataLoadedMsg struct {
	Todo     []data.Task
	Done     []data.Task
	Projects map[string]data.Project
}

func (a *AppModel) Init() tea.Cmd {
	a.stack = make([]tea.Model, 0)
	return func() tea.Msg {
		todo, done, projects, err := data.LoadData(false)
		if err != nil {
			if mismatchErr, ok := err.(*data.ParseTaskMismatchError); ok {
				return ParseTaskMismatchMsg{Err: mismatchErr}
			}
			return err // generic error message
		}
		return initDataLoadedMsg{todo, done, projects}
	}
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case initDataLoadedMsg:
		a.todoFileTasks = msg.Todo
		a.doneFileTasks = msg.Done
		a.projectMap = msg.Projects

		// Initialize and push TaskPicker
		taskPicker := components.NewTaskPickerModel(a.todoFileTasks)
		a.stack = append(a.stack, taskPicker)

		return a, taskPicker.Init()

	case ParseTaskMismatchMsg:
		// Handle the mismatch error here
		// For example, push a new error screen or print a message
		return a, tea.Printf("⚠️ Parse mismatch: %v", msg.Err)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	// If there’s a submodel, delegate updates to it
	if len(a.stack) > 0 {
		top := a.stack[len(a.stack)-1]
		newTop, cmd := top.Update(msg)
		a.stack[len(a.stack)-1] = newTop
		return a, cmd
	}

	return a, nil
}

func (a *AppModel) View() string {
	if len(a.stack) > 0 {
		return a.stack[len(a.stack)-1].View()
	}
	return "Home"
}
