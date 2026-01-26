package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wyattlefevre/wydocli/internal/data"
)

func TestTaskEditor_DueDateEdit(t *testing.T) {
	task := &data.Task{
		Name: "Test task",
		Tags: make(map[string]string),
	}

	editor := NewTaskEditor(task, []string{"proj1"}, []string{"ctx1"})

	// Press 'd' to enter due date edit mode
	model, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	editor = model.(*TaskEditorModel)

	if editor.inputContext.Mode != ModeEditDueDate {
		t.Errorf("expected ModeEditDueDate, got %v", editor.inputContext.Mode)
	}
	if editor.textInput == nil {
		t.Error("expected textInput to be created")
	}
	if cmd == nil {
		t.Error("expected focus command")
	}

	// Simulate entering a date
	editor.textInput.Input.SetValue("2025-01-15")

	// Press enter to confirm
	model, cmd = editor.Update(tea.KeyMsg{Type: tea.KeyEnter})
	editor = model.(*TaskEditorModel)

	// The textInput returns a command that sends TextInputResultMsg
	if cmd == nil {
		t.Error("expected command from enter")
	}

	// Simulate receiving the result message
	result := TextInputResultMsg{Value: "2025-01-15", Cancelled: false}
	model, _ = editor.Update(result)
	editor = model.(*TaskEditorModel)

	if editor.textInput != nil {
		t.Error("expected textInput to be nil after confirm")
	}
	if editor.inputContext.Mode != ModeTaskEditor {
		t.Errorf("expected ModeTaskEditor, got %v", editor.inputContext.Mode)
	}
	if task.GetDueDate() != "2025-01-15" {
		t.Errorf("expected due date '2025-01-15', got '%s'", task.GetDueDate())
	}
}

func TestTaskEditor_DueDateEdit_Cancel(t *testing.T) {
	task := &data.Task{
		Name: "Test task",
		Tags: map[string]string{"due": "2025-01-01"},
	}

	editor := NewTaskEditor(task, nil, nil)

	// Press 'd' to enter due date edit mode
	model, _ := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	editor = model.(*TaskEditorModel)

	// Press esc to cancel
	model, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyEsc})
	editor = model.(*TaskEditorModel)

	// The textInput returns a command that sends TextInputResultMsg
	if cmd == nil {
		t.Error("expected command from esc")
	}

	// Simulate receiving the cancelled result
	result := TextInputResultMsg{Value: "", Cancelled: true}
	model, _ = editor.Update(result)
	editor = model.(*TaskEditorModel)

	if editor.textInput != nil {
		t.Error("expected textInput to be nil after cancel")
	}
	if task.GetDueDate() != "2025-01-01" {
		t.Errorf("expected due date unchanged '2025-01-01', got '%s'", task.GetDueDate())
	}
}

func TestTaskEditor_ProjectEdit(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Projects: []string{"old_project"},
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, []string{"proj1", "proj2"}, nil)

	// Press 'p' to enter project edit mode
	model, _ := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	editor = model.(*TaskEditorModel)

	if editor.inputContext.Mode != ModeEditProject {
		t.Errorf("expected ModeEditProject, got %v", editor.inputContext.Mode)
	}
	if editor.fuzzyPicker == nil {
		t.Error("expected fuzzyPicker to be created")
	}
	if !editor.fuzzyPicker.MultiSelect {
		t.Error("expected fuzzy picker to have MultiSelect=true")
	}
	if !editor.fuzzyPicker.AllowCreate {
		t.Error("expected fuzzy picker to have AllowCreate=true")
	}

	// Simulate receiving the result message with new projects
	result := FuzzyPickerResultMsg{Selected: []string{"proj1", "proj2"}, Cancelled: false}
	model, _ = editor.Update(result)
	editor = model.(*TaskEditorModel)

	if editor.fuzzyPicker != nil {
		t.Error("expected fuzzyPicker to be nil after confirm")
	}
	if editor.inputContext.Mode != ModeTaskEditor {
		t.Errorf("expected ModeTaskEditor, got %v", editor.inputContext.Mode)
	}
	if len(task.Projects) != 2 || task.Projects[0] != "proj1" || task.Projects[1] != "proj2" {
		t.Errorf("expected projects [proj1, proj2], got %v", task.Projects)
	}
}

func TestFuzzyPicker_CreateNewInMultiSelect(t *testing.T) {
	picker := NewFuzzyPicker([]string{"existing1", "existing2"}, "Select Projects", true, true)

	// Type a new project name that doesn't exist
	picker.Query = "newproject"
	picker.filterItems()

	// Move cursor to the "Create new" option (after filtered items)
	picker.Cursor = len(picker.Filtered)

	// Toggle the "Create new" option with space
	picker.toggleCurrent()

	if !picker.Selected["newproject"] {
		t.Error("expected 'newproject' to be selected after toggle")
	}

	// Confirm and check the result
	cmd := picker.confirm()
	msg := cmd()
	result, ok := msg.(FuzzyPickerResultMsg)
	if !ok {
		t.Fatalf("expected FuzzyPickerResultMsg, got %T", msg)
	}

	found := false
	for _, s := range result.Selected {
		if s == "newproject" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'newproject' in selected, got %v", result.Selected)
	}
}

func TestFuzzyPicker_CreateNewWithExisting(t *testing.T) {
	picker := NewFuzzyPicker([]string{"existing1", "existing2"}, "Select Projects", true, true)

	// Select an existing item first
	picker.Cursor = 0
	picker.toggleCurrent()

	// Type a new project name
	picker.Query = "brandnew"
	picker.filterItems()

	// Move to "Create new" option and toggle it
	picker.Cursor = len(picker.Filtered)
	picker.toggleCurrent()

	// Confirm
	cmd := picker.confirm()
	msg := cmd()
	result := msg.(FuzzyPickerResultMsg)

	// Should have both the existing item and the new one
	if len(result.Selected) != 2 {
		t.Errorf("expected 2 selected items, got %d: %v", len(result.Selected), result.Selected)
	}

	hasExisting := false
	hasNew := false
	for _, s := range result.Selected {
		if s == "existing1" {
			hasExisting = true
		}
		if s == "brandnew" {
			hasNew = true
		}
	}
	if !hasExisting {
		t.Error("expected 'existing1' in selected")
	}
	if !hasNew {
		t.Error("expected 'brandnew' in selected")
	}
}

func TestTaskEditor_ContextEdit(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Contexts: []string{"old_ctx"},
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, []string{"home", "work"})

	// Press 't' to enter context edit mode
	model, _ := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	editor = model.(*TaskEditorModel)

	if editor.inputContext.Mode != ModeEditContext {
		t.Errorf("expected ModeEditContext, got %v", editor.inputContext.Mode)
	}
	if editor.fuzzyPicker == nil {
		t.Error("expected fuzzyPicker to be created")
	}

	// Simulate receiving the result message
	result := FuzzyPickerResultMsg{Selected: []string{"home"}, Cancelled: false}
	model, _ = editor.Update(result)
	editor = model.(*TaskEditorModel)

	if len(task.Contexts) != 1 || task.Contexts[0] != "home" {
		t.Errorf("expected contexts [home], got %v", task.Contexts)
	}
}

func TestTaskEditor_ContextEdit_WithCKey(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Contexts: []string{},
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, []string{"home"})

	// Press 'c' to enter context edit mode (alternative to 't')
	model, _ := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	editor = model.(*TaskEditorModel)

	if editor.inputContext.Mode != ModeEditContext {
		t.Errorf("expected ModeEditContext, got %v", editor.inputContext.Mode)
	}
	if editor.fuzzyPicker == nil {
		t.Error("expected fuzzyPicker to be created")
	}
}

func TestTaskEditor_PriorityCycle(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Priority: data.PriorityNone,
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, nil)

	// Starting from no priority, cycle through all priorities
	expectedPriorities := []data.Priority{
		data.PriorityA, data.PriorityB, data.PriorityC,
		data.PriorityD, data.PriorityE, data.PriorityF,
		data.PriorityNone,
	}

	for i, expected := range expectedPriorities {
		model, _ := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
		editor = model.(*TaskEditorModel)

		if task.Priority != expected {
			t.Errorf("cycle %d: expected priority %v, got %v", i+1, expected, task.Priority)
		}
	}
}

func TestTaskEditor_SaveAndClose(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Priority: data.PriorityA,
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, nil)

	// Press enter to save and close
	_, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command from enter")
	}

	// Execute the command and check the result
	msg := cmd()
	result, ok := msg.(TaskEditorResultMsg)
	if !ok {
		t.Fatalf("expected TaskEditorResultMsg, got %T", msg)
	}

	if result.Cancelled {
		t.Error("expected not cancelled")
	}
	if !result.Saved {
		t.Error("expected saved=true")
	}
	if result.Task.Name != "Test task" {
		t.Errorf("expected task name 'Test task', got '%s'", result.Task.Name)
	}
}

func TestTaskEditor_CancelAndRestore(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Priority: data.PriorityA,
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, nil)

	// Modify the task
	task.Priority = data.PriorityB

	// Press esc to cancel
	_, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected command from esc")
	}

	// Execute the command and check the result
	msg := cmd()
	result, ok := msg.(TaskEditorResultMsg)
	if !ok {
		t.Fatalf("expected TaskEditorResultMsg, got %T", msg)
	}

	if !result.Cancelled {
		t.Error("expected cancelled=true")
	}
	// Task should be restored to original
	if task.Priority != data.PriorityA {
		t.Errorf("expected priority restored to A, got %v", task.Priority)
	}
}

func TestTaskEditor_IsModified(t *testing.T) {
	task := &data.Task{
		Name:     "Test task",
		Priority: data.PriorityA,
		Tags:     make(map[string]string),
	}

	editor := NewTaskEditor(task, nil, nil)

	if editor.IsModified() {
		t.Error("expected not modified initially")
	}

	// Modify priority
	task.Priority = data.PriorityB
	if !editor.IsModified() {
		t.Error("expected modified after priority change")
	}

	// Restore
	task.Priority = data.PriorityA
	if editor.IsModified() {
		t.Error("expected not modified after restoration")
	}
}
