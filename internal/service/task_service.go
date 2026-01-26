package service

import (
	"fmt"
	"time"

	"github.com/wyattlefevre/wydocli/internal/data"
	"github.com/wyattlefevre/wydocli/logs"
)

// TaskService defines the interface for task operations.
// Both CLI and TUI use this interface to interact with tasks.
type TaskService interface {
	// List returns all tasks (both pending and done)
	List() ([]data.Task, error)

	// ListByProject returns tasks belonging to a specific project
	ListByProject(project string) ([]data.Task, error)

	// ListByContext returns tasks belonging to a specific context
	ListByContext(context string) ([]data.Task, error)

	// ListPending returns only incomplete tasks
	ListPending() ([]data.Task, error)

	// ListDone returns only completed tasks
	ListDone() ([]data.Task, error)

	// Get returns a single task by ID
	Get(id string) (*data.Task, error)

	// Add creates a new task from a raw todo.txt line
	Add(rawLine string) (*data.Task, error)

	// Update modifies an existing task
	Update(task data.Task) error

	// Complete marks a task as done
	Complete(id string) error

	// Delete removes a task by ID
	Delete(id string) error

	// Archive moves all completed tasks to done.txt
	Archive() error

	// GetProjects returns the project map
	GetProjects() map[string]data.Project

	// Reload refreshes the in-memory data from disk
	Reload() error
}

// taskServiceImpl is the concrete implementation of TaskService
type taskServiceImpl struct {
	tasks    []data.Task
	projects map[string]data.Project
}

// NewTaskService creates a new TaskService instance
func NewTaskService() (TaskService, error) {
	svc := &taskServiceImpl{}
	if err := svc.Reload(); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *taskServiceImpl) Reload() error {
	tasks, projects, err := data.LoadData(true)
	if err != nil {
		return err
	}
	s.tasks = tasks
	s.projects = projects
	return nil
}

func (s *taskServiceImpl) List() ([]data.Task, error) {
	return s.tasks, nil
}

func (s *taskServiceImpl) ListByProject(project string) ([]data.Task, error) {
	var filtered []data.Task
	for _, t := range s.tasks {
		if t.HasProject(project) {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (s *taskServiceImpl) ListByContext(context string) ([]data.Task, error) {
	var filtered []data.Task
	for _, t := range s.tasks {
		if t.HasContext(context) {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (s *taskServiceImpl) ListPending() ([]data.Task, error) {
	var pending []data.Task
	for _, t := range s.tasks {
		if !t.Done {
			pending = append(pending, t)
		}
	}
	return pending, nil
}

func (s *taskServiceImpl) ListDone() ([]data.Task, error) {
	var done []data.Task
	for _, t := range s.tasks {
		if t.Done {
			done = append(done, t)
		}
	}
	return done, nil
}

func (s *taskServiceImpl) Get(id string) (*data.Task, error) {
	for _, t := range s.tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task not found: %s", id)
}

func (s *taskServiceImpl) Add(rawLine string) (*data.Task, error) {
	task, err := data.AppendTask(rawLine)
	if err != nil {
		return nil, err
	}
	// Reload to get fresh state
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskServiceImpl) Update(task data.Task) error {
	logs.Logger.Printf("Service: Update Task: %s\n", task.ID)
	data.UpdateTask(s.tasks, task)
	if err := data.WriteData(s.tasks); err != nil {
		return err
	}
	return s.Reload()
}

func (s *taskServiceImpl) Complete(id string) error {
	task, err := s.Get(id)
	if err != nil {
		return err
	}

	task.Done = true
	task.CompletionDate = time.Now().Format("2006-01-02")
	task.File = data.GetDoneFilePath()

	data.UpdateTask(s.tasks, *task)
	if err := data.WriteData(s.tasks); err != nil {
		return err
	}
	return s.Reload()
}

func (s *taskServiceImpl) Delete(id string) error {
	s.tasks = data.DeleteTask(s.tasks, id)
	if err := data.WriteData(s.tasks); err != nil {
		return err
	}
	return s.Reload()
}

func (s *taskServiceImpl) Archive() error {
	if err := data.ArchiveDone(s.tasks); err != nil {
		return err
	}
	return s.Reload()
}

func (s *taskServiceImpl) GetProjects() map[string]data.Project {
	return s.projects
}
