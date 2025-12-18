package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/wyattlefevre/wydocli/internal/data"
)

var (
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	priorityStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	projectStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	contextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	tagStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	nameStyle     = lipgloss.NewStyle().Bold(true)
	dateStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

func StyledTaskLine(t data.Task) string {
	taskLine := []string{}
	if t.Done {
		return renderDone(t)
	}
	taskLine = append(taskLine, doneStyle.Render("[ ] "))

	if t.Priority != 0 {
		taskLine = append(taskLine, priorityStyle.Render("("+string(t.Priority)+") "))
	}
	if t.CreatedDate != "" {
		taskLine = append(taskLine, dateStyle.Render(t.CreatedDate))
	}
	if t.CompletionDate != "" {
		taskLine = append(taskLine, dateStyle.Render(t.CompletionDate))
	}
	if t.Name != "" {
		taskLine = append(taskLine, nameStyle.Render(t.Name))
	}
	for _, p := range t.Projects {
		taskLine = append(taskLine, projectStyle.Render(" +"+p))
	}
	for _, c := range t.Contexts {
		taskLine = append(taskLine, contextStyle.Render(" @"+c))
	}
	for k, v := range t.Tags {
		taskLine = append(taskLine, tagStyle.Render(" "+k+":"+v))
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, taskLine...)
	return lipgloss.NewStyle().Padding(0, 1).Render(line)
}

func renderDone(t data.Task) string {
	taskLine := []string{}
	taskLine = append(taskLine, doneStyle.Render("[x] "))
	if t.Priority != 0 {
		taskLine = append(taskLine, doneStyle.Render("("+string(t.Priority)+")"))
	}
	if t.Priority != 0 {
		taskLine = append(taskLine, doneStyle.Render("("+string(t.Priority)+")"))
	}
	if t.CreatedDate != "" {
		taskLine = append(taskLine, doneStyle.Render(t.CreatedDate))
	}
	if t.CompletionDate != "" {
		taskLine = append(taskLine, doneStyle.Render(t.CompletionDate))
	}
	if t.Name != "" {
		taskLine = append(taskLine, doneStyle.Render(t.Name))
	}
	for _, p := range t.Projects {
		taskLine = append(taskLine, doneStyle.Render("+"+p))
	}
	for _, c := range t.Contexts {
		taskLine = append(taskLine, doneStyle.Render("@"+c))
	}
	for k, v := range t.Tags {
		taskLine = append(taskLine, doneStyle.Render(k+":"+v))
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, taskLine...)
	return lipgloss.NewStyle().Padding(0, 1).Render(line)
}
