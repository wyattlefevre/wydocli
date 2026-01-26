package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFuzzyPicker_StartsInNavigationMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)
	if picker.filterMode {
		t.Error("expected picker to start in navigation mode")
	}
}

func TestFuzzyPicker_SlashEntersFilterMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Press "/"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(msg)

	if !picker.filterMode {
		t.Error("expected '/' to enter filter mode")
	}
}

func TestFuzzyPicker_JKNavigateInNavigationMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Press "j" - should move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	picker.Update(msg)
	if picker.Cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", picker.Cursor)
	}

	// Press "k" - should move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	picker.Update(msg)
	if picker.Cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", picker.Cursor)
	}
}

func TestFuzzyPicker_JKTypeInFilterMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)

	// Type "jk"
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	picker.Update(jMsg)
	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	picker.Update(kMsg)

	// Check that j/k were typed, not navigation
	if picker.Query != "jk" {
		t.Errorf("expected query 'jk', got '%s'", picker.Query)
	}
	if picker.Cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", picker.Cursor)
	}
}

func TestFuzzyPicker_EnterExitsFilterMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode and type
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	picker.Update(aMsg)

	// Press enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	picker.Update(enterMsg)

	if picker.filterMode {
		t.Error("expected enter to exit filter mode")
	}
	if picker.Query != "a" {
		t.Errorf("expected query 'a' to be preserved, got '%s'", picker.Query)
	}
}

func TestFuzzyPicker_EscClearsQueryInFilterMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode and type
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	picker.Update(aMsg)

	if picker.Query != "a" {
		t.Fatalf("expected query 'a', got '%s'", picker.Query)
	}

	// Press esc
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	picker.Update(escMsg)

	if picker.filterMode {
		t.Error("expected esc to exit filter mode")
	}
	if picker.Query != "" {
		t.Errorf("expected query to be cleared, got '%s'", picker.Query)
	}
}

func TestFuzzyPicker_EscClearsQueryInNavigationMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode, type, then exit with enter
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	picker.Update(aMsg)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	picker.Update(enterMsg)

	// Now in navigation mode with query
	if picker.Query != "a" {
		t.Fatalf("expected query 'a', got '%s'", picker.Query)
	}

	// Press esc - should clear query, not cancel
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	result, cmd := picker.Update(escMsg)
	_ = result

	if picker.Query != "" {
		t.Errorf("expected query to be cleared, got '%s'", picker.Query)
	}
	if cmd != nil {
		t.Error("expected no command (not cancelled) when clearing query")
	}
}

func TestFuzzyPicker_DoubleEscCancelsPicker(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode, type, then exit with enter
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	picker.Update(aMsg)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	picker.Update(enterMsg)

	// First esc clears query
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	picker.Update(escMsg)

	// Second esc should cancel
	_, cmd := picker.Update(escMsg)

	if cmd == nil {
		t.Fatal("expected command when cancelling")
	}

	// Execute the command to get the message
	msg := cmd()
	resultMsg, ok := msg.(FuzzyPickerResultMsg)
	if !ok {
		t.Fatalf("expected FuzzyPickerResultMsg, got %T", msg)
	}
	if !resultMsg.Cancelled {
		t.Error("expected picker to be cancelled")
	}
}

func TestFuzzyPicker_SpaceTogglesInNavigationMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", true, false)

	// Press space to toggle first item
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	picker.Update(spaceMsg)

	if !picker.Selected["alpha"] {
		t.Error("expected 'alpha' to be selected")
	}

	// Toggle again
	picker.Update(spaceMsg)
	if picker.Selected["alpha"] {
		t.Error("expected 'alpha' to be deselected")
	}
}

func TestFuzzyPicker_FilteringWorks(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode and type "al"
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)
	aMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	picker.Update(aMsg)
	lMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	picker.Update(lMsg)

	if len(picker.Filtered) != 1 {
		t.Errorf("expected 1 filtered item, got %d", len(picker.Filtered))
	}
	if picker.Filtered[0] != "alpha" {
		t.Errorf("expected 'alpha', got '%s'", picker.Filtered[0])
	}
}

func TestFuzzyPicker_NavigateAfterFilterMode(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta", "gamma"}, "Test", false, false)

	// Enter filter mode
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)

	// Exit filter mode with enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	picker.Update(enterMsg)

	// Now navigate with j/k
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	picker.Update(jMsg)

	if picker.Cursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", picker.Cursor)
	}
}

func TestFuzzyPicker_ViewShowsCorrectHelpText(t *testing.T) {
	picker := NewFuzzyPicker([]string{"alpha", "beta"}, "Test", false, false)

	// Navigation mode - should show filter help
	view := picker.View()
	if !containsString(view, "[/] filter") {
		t.Error("expected navigation mode view to show [/] filter")
	}

	// Enter filter mode
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	picker.Update(slashMsg)

	view = picker.View()
	if !containsString(view, "[enter] done") {
		t.Error("expected filter mode view to show [enter] done")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
