package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
)

func TestTaskManager_ForwardsTextInputResultToTaskEditor(t *testing.T) {
	// Create a task manager with a task
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "Test task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Open the task editor
	tm.cursor = 0
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	if tm.taskEditor == nil {
		t.Fatal("expected task editor to be open")
	}

	// Press 'd' to open the date input
	model, _ = tm.taskEditor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	tm.taskEditor = model.(*TaskEditorModel)

	if tm.taskEditor.textInput == nil {
		t.Fatal("expected text input to be open in task editor")
	}

	// Simulate a TextInputResultMsg - this should be forwarded to the task editor
	// not handled by the task manager
	result := TextInputResultMsg{Value: "2025-06-15", Cancelled: false}
	model, _ = tm.Update(result)
	tm = model.(*TaskManagerModel)

	// Task editor should still be open (not closed by the task manager's handler)
	if tm.taskEditor == nil {
		t.Error("task editor should still be open after date input result")
	}

	// The text input in the task editor should be closed
	if tm.taskEditor.textInput != nil {
		t.Error("task editor's text input should be nil after result")
	}

	// The task's due date should be updated
	if tm.displayTasks[0].GetDueDate() != "2025-06-15" {
		t.Errorf("expected due date '2025-06-15', got '%s'", tm.displayTasks[0].GetDueDate())
	}
}

func TestTaskManager_ForwardsFuzzyPickerResultToTaskEditor(t *testing.T) {
	// Create a task manager with a task
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "Test task", Projects: []string{}, Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)
	tm.allProjects = []string{"proj1", "proj2"}

	// Open the task editor
	tm.cursor = 0
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	if tm.taskEditor == nil {
		t.Fatal("expected task editor to be open")
	}

	// Press 'p' to open the project picker
	model, _ = tm.taskEditor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	tm.taskEditor = model.(*TaskEditorModel)

	if tm.taskEditor.fuzzyPicker == nil {
		t.Fatal("expected fuzzy picker to be open in task editor")
	}

	// Simulate a FuzzyPickerResultMsg - this should be forwarded to the task editor
	result := FuzzyPickerResultMsg{Selected: []string{"proj1"}, Cancelled: false}
	model, _ = tm.Update(result)
	tm = model.(*TaskManagerModel)

	// Task editor should still be open
	if tm.taskEditor == nil {
		t.Error("task editor should still be open after project selection")
	}

	// The fuzzy picker in the task editor should be closed
	if tm.taskEditor.fuzzyPicker != nil {
		t.Error("task editor's fuzzy picker should be nil after result")
	}

	// The task's projects should be updated
	if len(tm.displayTasks[0].Projects) != 1 || tm.displayTasks[0].Projects[0] != "proj1" {
		t.Errorf("expected projects [proj1], got %v", tm.displayTasks[0].Projects)
	}
}

func TestTaskManager_HandlesOwnTextInputResult(t *testing.T) {
	// Create a task manager
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "Test task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// When task editor is NOT open and task manager's own text input is used
	// the task manager should handle the result itself
	tm.textInput = NewDateInput("Filter Date")
	tm.inputContext.Mode = ModeSearch

	result := TextInputResultMsg{Value: "test query", Cancelled: false}
	model, _ := tm.Update(result)
	tm = model.(*TaskManagerModel)

	// Text input should be cleared
	if tm.textInput != nil {
		t.Error("expected text input to be nil after result")
	}
}

func TestTaskManager_HandlesOwnFuzzyPickerResult(t *testing.T) {
	// Create a task manager
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "Test task", Projects: []string{"proj1"}, Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)
	tm.allProjects = []string{"proj1", "proj2"}

	// When task editor is NOT open and task manager's own fuzzy picker is used
	tm.fuzzyPicker = NewFuzzyPicker(tm.allProjects, "Filter", true, false)
	tm.pickerContext = "filter-project"
	tm.inputContext.Mode = ModeFuzzyPicker

	result := FuzzyPickerResultMsg{Selected: []string{"proj1"}, Cancelled: false}
	model, _ := tm.Update(result)
	tm = model.(*TaskManagerModel)

	// Fuzzy picker should be cleared
	if tm.fuzzyPicker != nil {
		t.Error("expected fuzzy picker to be nil after result")
	}

	// Filter should be applied
	if len(tm.filterState.ProjectFilter) != 1 || tm.filterState.ProjectFilter[0] != "proj1" {
		t.Errorf("expected project filter [proj1], got %v", tm.filterState.ProjectFilter)
	}
}

func TestTaskManager_TaskEditorCloseReturnsToTaskList(t *testing.T) {
	// Create a task manager with tasks
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "Test task", Priority: data.PriorityA, Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Open the task editor
	tm.cursor = 0
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	if tm.taskEditor == nil {
		t.Fatal("expected task editor to be open")
	}

	// Press enter in task editor to save and close
	_, cmd := tm.taskEditor.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from enter")
	}

	// Execute the command to get TaskEditorResultMsg
	msg := cmd()
	result, ok := msg.(TaskEditorResultMsg)
	if !ok {
		t.Fatalf("expected TaskEditorResultMsg, got %T", msg)
	}

	// Handle the result in task manager
	model, _ = tm.Update(result)
	tm = model.(*TaskManagerModel)

	// Task editor should be closed
	if tm.taskEditor != nil {
		t.Error("expected task editor to be nil after save")
	}

	// Mode should be reset to normal
	if tm.inputContext.Mode != ModeNormal {
		t.Errorf("expected ModeNormal, got %v", tm.inputContext.Mode)
	}
}

// Search filter mode tests

func TestTaskManager_SearchStartsInFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
		{Name: "beta task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Press "/" to start search
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	if !tm.searchActive {
		t.Error("expected search to be active")
	}
	if !tm.searchFilterMode {
		t.Error("expected search to start in filter mode")
	}
}

func TestTaskManager_SearchJKTypeInFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
		{Name: "beta task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
		{Name: "gamma task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search (enters filter mode)
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	// Type "jk" - should go to input, not navigate
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	tm = model.(*TaskManagerModel)
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	tm = model.(*TaskManagerModel)

	if tm.filterState.SearchQuery != "jk" {
		t.Errorf("expected query 'jk', got '%s'", tm.filterState.SearchQuery)
	}
	if tm.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", tm.cursor)
	}
}

func TestTaskManager_SearchEnterExitsFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	// Type something
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm = model.(*TaskManagerModel)

	// Press enter - should exit filter mode but stay in search mode
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	if tm.searchFilterMode {
		t.Error("expected filter mode to be false after enter")
	}
	if !tm.searchActive {
		t.Error("expected search to still be active")
	}
	if tm.filterState.SearchQuery != "a" {
		t.Errorf("expected query 'a' to be preserved, got '%s'", tm.filterState.SearchQuery)
	}
}

func TestTaskManager_SearchJKNavigateAfterExitingFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
		{Name: "beta task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
		{Name: "gamma task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	// Exit filter mode with enter
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	// Now j/k should navigate
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	tm = model.(*TaskManagerModel)

	if tm.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", tm.cursor)
	}

	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	tm = model.(*TaskManagerModel)

	if tm.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", tm.cursor)
	}
}

func TestTaskManager_SearchSlashReEntersFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	// Exit filter mode with enter
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	if tm.searchFilterMode {
		t.Fatal("expected filter mode to be false")
	}

	// Press "/" to re-enter filter mode
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)

	if !tm.searchFilterMode {
		t.Error("expected '/' to re-enter filter mode")
	}
}

func TestTaskManager_SearchEscClearsQueryInFilterMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search and type
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm = model.(*TaskManagerModel)

	if tm.filterState.SearchQuery != "a" {
		t.Fatalf("expected query 'a', got '%s'", tm.filterState.SearchQuery)
	}

	// Press esc - should clear query and exit filter mode
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEscape})
	tm = model.(*TaskManagerModel)

	if tm.searchFilterMode {
		t.Error("expected filter mode to be false after esc")
	}
	if tm.filterState.SearchQuery != "" {
		t.Errorf("expected query to be cleared, got '%s'", tm.filterState.SearchQuery)
	}
	if !tm.searchActive {
		t.Error("expected search to still be active")
	}
}

func TestTaskManager_SearchDoubleEscExitsSearchMode(t *testing.T) {
	tm := &TaskManagerModel{}
	tm.Init()
	tasks := []data.Task{
		{Name: "alpha task", Tags: make(map[string]string), File: data.GetTodoFilePath()},
	}
	tm.WithTasks(tasks)

	// Start search, type, exit filter mode
	model, _ := tm.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	tm = model.(*TaskManagerModel)
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm = model.(*TaskManagerModel)
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEnter})
	tm = model.(*TaskManagerModel)

	// First esc clears query
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEscape})
	tm = model.(*TaskManagerModel)

	if tm.filterState.SearchQuery != "" {
		t.Errorf("expected query to be cleared, got '%s'", tm.filterState.SearchQuery)
	}
	if !tm.searchActive {
		t.Error("expected search to still be active after first esc")
	}

	// Second esc exits search mode
	model, _ = tm.handleSearchMode(tea.KeyMsg{Type: tea.KeyEscape})
	tm = model.(*TaskManagerModel)

	if tm.searchActive {
		t.Error("expected search mode to be exited after second esc")
	}
}
