package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wyattlefevre/wydocli/internal/components"
	"github.com/wyattlefevre/wydocli/internal/data"
)

type AppModel struct {
	taskManager    tea.Model
	projectManager tea.Model
	currentView    ViewType
	tasks          []data.Task
	projects       map[string]data.Project
	loading        bool
}

type ViewType int

const (
	ViewTaskManager ViewType = iota
	ViewProjectManager
	ViewFileManager
)

type ParseTaskMismatchMsg struct {
	Err *data.ParseTaskMismatchError
}

type DataLoadedMsg struct {
	Tasks    []data.Task
	Projects map[string]data.Project
}

func NewAppModel() *AppModel {
	return &AppModel{
		taskManager:    &components.TaskManagerModel{},
		projectManager: &components.ProjectManagerModel{},
		currentView:    ViewTaskManager, // or whichever view you want to start with
		tasks:          make([]data.Task, 0),
		projects:       make(map[string]data.Project),
		loading:        false,
	}
}

func (a *AppModel) Init() tea.Cmd {
	return func() tea.Msg {
		a.loading = true
		tasks, projects, err := data.LoadData(false)
		if err != nil {
			if mismatchErr, ok := err.(*data.ParseTaskMismatchError); ok {
				return ParseTaskMismatchMsg{Err: mismatchErr}
			}
			return err // generic error message
		}
		return DataLoadedMsg{tasks, projects}
	}
}

func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case DataLoadedMsg:
		a.tasks = msg.Tasks
		a.projects = msg.Projects
		a.loading = false

		if tm, ok := a.taskManager.(*components.TaskManagerModel); ok {
			a.taskManager = tm.WithTasks(a.tasks)
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
		case "P":
			a.currentView = ViewProjectManager
			return a, nil
		case "T":
			a.currentView = ViewTaskManager
			return a, nil
		case "F":
			// TODO implement file view
			return a, nil
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
			return DataLoadedMsg{tasks, projects}
		}

	}

	var cmd tea.Cmd
	switch a.currentView {
	case ViewTaskManager:
		a.taskManager, cmd = a.taskManager.Update(msg)
	case ViewProjectManager:
		a.projectManager, cmd = a.projectManager.Update(msg)
	}

	return a, cmd
}

func (a *AppModel) View() string {
	topBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Padding(0, 1).
		Bold(true)

	topBar := topBarStyle.Render(" WYDO CLI | [P] Projects | [T] Tasks | [F] Files | [q] Quit")
	var b strings.Builder
	content := ""
	switch a.currentView {
	case ViewTaskManager:
		content = a.taskManager.View()
	case ViewProjectManager:
		content = a.projectManager.View()
	}
	b.WriteString(topBar)
	b.WriteString("\n\n")
	b.WriteString(content)
	return b.String()
}
