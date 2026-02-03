package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wyattlefevre/wydocli/internal/data"
)

var (
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	priorityStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	projectStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	contextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	tagStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	nameStyle     = lipgloss.NewStyle()
	dateStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

// StyledTaskLine renders a task in a simple, readable format.
// Format: [x] (A) Name +project @context due:date
func StyledTaskLine(t data.Task) string {
	var parts []string

	// Status checkbox
	if t.Done {
		parts = append(parts, doneStyle.Render("[x]"))
	} else {
		parts = append(parts, "[ ]")
	}

	// Priority
	if t.Priority != 0 {
		parts = append(parts, priorityStyle.Render("("+string(t.Priority)+")"))
	}
	if t.CreatedDate != "" {
		parts = append(parts, dateStyle.Render(t.CreatedDate))
	}
	if t.CompletionDate != "" {
		parts = append(parts, dateStyle.Render(t.CompletionDate))
	}

	// Name
	if t.Name != "" {
		if t.Done {
			parts = append(parts, doneStyle.Render(t.Name))
		} else {
			parts = append(parts, nameStyle.Render(t.Name))
		}
	}

	// Projects
	for _, p := range t.Projects {
		parts = append(parts, projectStyle.Render("+"+p))
	}

	// Contexts
	for _, c := range t.Contexts {
		parts = append(parts, contextStyle.Render("@"+c))
	}

	// Tags (including due date)
	for k, v := range t.Tags {
		parts = append(parts, tagStyle.Render(k+":"+v))
	}

	return strings.Join(parts, " ")
}
