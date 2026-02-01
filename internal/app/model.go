package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wyattlefevre/wydocli/internal/components"
	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/internal/service"
	"github.com/wyattlefevre/wydocli/logs"
)

type AppModel struct {
	taskManager    tea.Model
	projectManager tea.Model
	currentView    ViewType
	tasks          []data.Task
	projects       map[string]data.Project
	loading        bool
	service        service.TaskService
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

// NewAppModel creates a new AppModel without a service (legacy, loads data internally)
func NewAppModel() *AppModel {
	return &AppModel{
		taskManager:    &components.TaskManagerModel{},
		projectManager: &components.ProjectManagerModel{},
		currentView:    ViewTaskManager,
		tasks:          make([]data.Task, 0),
		projects:       make(map[string]data.Project),
		loading:        false,
		service:        nil,
	}
}

// NewAppModelWithService creates a new AppModel with an injected TaskService
func NewAppModelWithService(svc service.TaskService) *AppModel {
	return &AppModel{
		taskManager:    &components.TaskManagerModel{},
		projectManager: &components.ProjectManagerModel{},
		currentView:    ViewTaskManager,
		tasks:          make([]data.Task, 0),
		projects:       make(map[string]data.Project),
		loading:        false,
		service:        svc,
	}
}

func (a *AppModel) Init() tea.Cmd {
	return func() tea.Msg {
		a.loading = true

		var tasks []data.Task
		var projects map[string]data.Project
		var err error

		if a.service != nil {
			// Use service if available
			tasks, err = a.service.List()
			if err != nil {
				logs.Logger.Fatalf("ERROR: %s", err.Error())
				return err
			}
			projects = a.service.GetProjects()
		} else {
			// Fallback to direct data loading (legacy)
			tasks, projects, err = data.LoadData(true)
			if err != nil {
				logs.Logger.Fatalf("ERROR: %s", err.Error())
				if mismatchErr, ok := err.(*data.ParseTaskMismatchError); ok {
					return ParseTaskMismatchMsg{Err: mismatchErr}
				}
				return err
			}
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
		logs.Logger.Println("Parse Mismatch detected, must resolve")
		return a, tea.Printf("⚠️ Parse mismatch: %v", msg.Err)

	case tea.KeyMsg:
		if a.loading {
			return a, nil
		}

		// Check if task manager is in a modal state (editor, picker, input, or non-normal mode)
		// If so, pass keys to task manager first
		if tm, ok := a.taskManager.(*components.TaskManagerModel); ok && tm.IsInModalState() {
			var cmd tea.Cmd
			a.taskManager, cmd = a.taskManager.Update(msg)
			return a, cmd
		}

		// Global keys only when not in modal state
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
			// Shift+F - Toggle file view
			if _, ok := a.taskManager.(*components.TaskManagerModel); ok {
				return a, func() tea.Msg {
					return components.ToggleFileViewMsg{}
				}
			}
			return a, nil
		case "A":
			// Shift+A - Archive completed tasks
			if _, ok := a.taskManager.(*components.TaskManagerModel); ok {
				return a, func() tea.Msg {
					return components.StartArchiveMsg{}
				}
			}
			return a, nil
		}

	case components.TaskUpdateMsg:
		// Update the task using service or data layer
		a.loading = true

		if a.service != nil {
			return a, func() tea.Msg {
				err := a.service.Update(msg.Task)
				if err != nil {
					return tea.Printf("Error updating task: %v", err)
				}
				tasks, err := a.service.List()
				if err != nil {
					return tea.Printf("Error loading tasks: %v", err)
				}
				return DataLoadedMsg{tasks, a.service.GetProjects()}
			}
		}

		// Legacy path without service
		a.tasks = data.UpdateTask(a.tasks, msg.Task)
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

	case components.ArchiveRequestMsg:
		a.loading = true
		count := msg.Count
		return a, func() tea.Msg {
			if a.service != nil {
				err := a.service.Archive()
				if err != nil {
					return tea.Printf("Error archiving: %v", err)
				}
				tasks, err := a.service.List()
				if err != nil {
					return tea.Printf("Error loading: %v", err)
				}
				a.tasks = tasks
				return components.ArchiveCompleteMsg{Count: count}
			}

			// Legacy path without service
			err := data.ArchiveDone(a.tasks)
			if err != nil {
				return tea.Printf("Error archiving: %v", err)
			}
			tasks, projects, err := data.LoadData(false)
			if err != nil {
				return tea.Printf("Error loading: %v", err)
			}
			a.tasks = tasks
			a.projects = projects
			return components.ArchiveCompleteMsg{Count: count}
		}

	case components.ArchiveCompleteMsg:
		a.loading = false
		if tm, ok := a.taskManager.(*components.TaskManagerModel); ok {
			a.taskManager = tm.WithTasks(a.tasks)
		}
		// Forward message to task manager for success display
		var cmd tea.Cmd
		a.taskManager, cmd = a.taskManager.Update(msg)
		return a, cmd
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
