package data

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type Priority rune

const (
	PriorityA    Priority = 'A'
	PriorityB    Priority = 'B'
	PriorityC    Priority = 'C'
	PriorityD    Priority = 'D'
	PriorityE    Priority = 'E'
	PriorityF    Priority = 'F'
	PriorityNone Priority = 0
)

type Task struct {
	ID             string
	Name           string
	Projects       []string
	Contexts       []string
	Done           bool
	Tags           map[string]string
	CreatedDate    string
	CompletionDate string
	Priority       Priority
	File           string
	DueDate        string
}

func (t *Task) HasProject(project string) bool {
	return slices.Contains(t.Projects, project)
}

func (t *Task) AddProject(project string) {
	if !t.HasProject(project) {
		t.Projects = append(t.Projects, project)
	}
}

func (t *Task) RemoveProject(project string) {
	for i, p := range t.Projects {
		if p == project {
			t.Projects = append(t.Projects[:i], t.Projects[i+1:]...)
			break
		}
	}
}

func (t *Task) HasContext(context string) bool {
	return slices.Contains(t.Contexts, context)
}

func (t *Task) AddContext(context string) {
	if !t.HasContext(context) {
		t.Contexts = append(t.Contexts, context)
	}
}

func (t *Task) RemoveContext(context string) {
	for i, c := range t.Contexts {
		if c == context {
			t.Contexts = append(t.Contexts[:i], t.Contexts[i+1:]...)
			break
		}
	}
}

func (t Task) String() string {
	var parts []string

	// Done status
	if t.Done {
		parts = append(parts, "x")
	}

	// Priority
	if t.Priority != 0 {
		parts = append(parts, "("+string(t.Priority)+")")
	}

	// Dates
	if t.CreatedDate != "" {
		parts = append(parts, t.CreatedDate)
	}
	if t.CompletionDate != "" {
		parts = append(parts, t.CompletionDate)
	}

	// Name
	if t.Name != "" {
		parts = append(parts, t.Name)
	}

	// Projects
	for _, p := range t.Projects {
		parts = append(parts, "+"+p)
	}

	// Contexts
	for _, c := range t.Contexts {
		parts = append(parts, "@"+c)
	}

	// Tags
	for k, v := range t.Tags {
		parts = append(parts, k+":"+v)
	}

	return strings.Join(parts, " ")
}

func (t Task) Print() {
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Name: %s\n", t.Name)
	fmt.Printf("Projects: %v\n", t.Projects)
	fmt.Printf("Contexts: %v\n", t.Contexts)
	fmt.Printf("Done: %v\n", t.Done)
	fmt.Printf("Tags: %v\n", t.Tags)
	fmt.Printf("CreatedDate: %s\n", t.CreatedDate)
	fmt.Printf("CompletionDate: %s\n", t.CompletionDate)
	fmt.Printf("Priority: %c\n", t.Priority)
}

func ParseTask(input string, id string, file string) Task {
	input = strings.TrimSpace(input)
	input = CollapseWhitespace(input)

	var t Task
	t.ID = id
	t.File = file

	if strings.HasPrefix(input, "x ") {
		t.Done = true
		input = input[2:]
	}

	t.Priority = ParsePriority(input)
	if t.Priority != PriorityNone {
		input = input[3:]
	}

	firstMetaIdx := FirstMetaIndex(
		FirstProjectIndex(input),
		FirstContextIndex(input),
		FirstTagIndex(input),
	)

	if firstMetaIdx == -1 {
		t.Name = input
		// task has no metadata (project, context, tag)
		return t
	}

	return t
}

func CollapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func FirstProjectIndex(s string) int {
	re := regexp.MustCompile(`[ \t]\+[A-Za-z0-9]`)
	loc := re.FindStringIndex(s)
	if loc != nil {
		// Return the index of the "+" character
		return loc[0] + 1
	}
	return -1
}

func FirstContextIndex(s string) int {
	re := regexp.MustCompile(`[ \t]\@[A-Za-z0-9]`)
	loc := re.FindStringIndex(s)
	if loc != nil {
		// Return the index of the "@" character
		return loc[0] + 1
	}
	return -1
}

func FirstTagIndex(s string) int {
	re := regexp.MustCompile(`[ \t][A-Za-z0-9]+\:[A-Za-z0-9]+`)
	loc := re.FindStringIndex(s)
	if loc != nil {
		// Return the index of the first character of the tag (after the space or tab)
		return loc[0] + 1
	}
	return -1
}

func ParseProjects(s string) []string {
	re := regexp.MustCompile(`[ \t]\+[A-Za-z0-9]+`)
	matches := re.FindAllString(s, -1)
	for i, m := range matches {
		matches[i] = m[2:]
	}
	return matches
}

func ParseContexts(s string) []string {
	re := regexp.MustCompile(`[ \t]\@[A-Za-z0-9]+`)
	matches := re.FindAllString(s, -1)
	for i, m := range matches {
		matches[i] = m[2:]
	}
	return matches
}

func ParseTags(s string) []string {
	re := regexp.MustCompile(`[ \t][A-Za-z0-9]+\:[A-Za-z0-9]+`)
	matches := re.FindAllString(s, -1)
	trimmed := make([]string, len(matches))
	for i, m := range matches {
		trimmed[i] = strings.TrimSpace(m)
	}
	return trimmed
}

func ParsePriority(s string) Priority {
	re := regexp.MustCompile(`^\(([A-Fa-f])\)`)
	matches := re.FindStringSubmatch(s)
	if matches != nil {
		switch strings.ToUpper(matches[1]) {
		case "A":
			return PriorityA
		case "B":
			return PriorityB
		case "C":
			return PriorityC
		case "D":
			return PriorityD
		case "E":
			return PriorityE
		case "F":
			return PriorityF
		}
	}
	return PriorityNone
}

func FirstMetaIndex(i1 int, i2 int, i3 int) int {
	min := -1
	for _, v := range []int{i1, i2, i3} {
		if v >= 0 && (min == -1 || v < min) {
			min = v
		}
	}
	return min
}
