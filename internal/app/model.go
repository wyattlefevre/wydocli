package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/components"
	"github.com/wyattlefevre/wydocli/internal/data"
)

type AppModel struct {
	stack    []tea.Model // TODO: might not need a stack. would make it easier to explicitly update and control components. in fact, it would be nice for preserving state. Only need one task picker loaded
	tasks    []data.Task
	projects map[string]data.Project
	loading  bool
}

type ParseTaskMismatchMsg struct {
	Err *data.ParseTaskMismatchError
}

type dataLoadedMsg struct {
	Tasks    []data.Task
	Projects map[string]data.Project
}

func (a *AppModel) Init() tea.Cmd {
	a.stack = make([]tea.Model, 0)
	return func() tea.Msg {
		tasks, projects, err := data.LoadData(false)
		if err != nil {
			if mismatchErr, ok := err.(*data.ParseTaskMismatchError); ok {
				return ParseTaskMismatchMsg{Err: mismatchErr}
			}
			return err // generic error message
		}
		return dataLoadedMsg{tasks, projects}
	}
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case dataLoadedMsg:
		a.tasks = msg.Tasks
		a.projects = msg.Projects
		a.loading = false

		// Initialize and push TaskPicker
		if len(a.stack) == 0 {
			taskPicker := components.NewTaskPickerModel(a.tasks)
			a.stack = append(a.stack, taskPicker)
			return a, taskPicker.Init()
		}

		return a, nil

	case ParseTaskMismatchMsg:
		// Handle the mismatch error here
		// For example, push a new error screen or print a message
		return a, tea.Printf("⚠️ Parse mismatch: %v", msg.Err)

	case tea.KeyMsg:
		if a.loading {
			return a, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

	case components.TaskUpdateMsg:
		// Update the task in a.tasks
		data.UpdateTask(a.tasks, msg.Task)

		// Block input while loading
		a.loading = true
		// Write to disk and then reload data in a Cmd
		return a, func() tea.Msg {
			err := data.WriteData(a.tasks)
			if err != nil {
				return tea.Printf("Error writing tasks: %v", err)
			}
			tasks, projects, err := data.LoadData(false)
			if err != nil {
				return tea.Printf("Error loading tasks: %v", err)
			}
			return dataLoadedMsg{tasks, projects}
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
