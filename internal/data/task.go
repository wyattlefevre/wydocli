package data

import (
	"fmt"
	"regexp"
	"strings"
)

type Priority rune

const (
	PriorityA Priority = 'A'
	PriorityB Priority = 'B'
	PriorityC Priority = 'C'
	PriorityD Priority = 'D'
	PriorityE Priority = 'E'
	PriorityF Priority = 'F'
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
}

func ParseTask(input string, id string) Task {
	var t Task
	t.ID = id

	// Completion: starts with "x "
	if strings.HasPrefix(input, "x ") {
		t.Done = true
		input = input[2:]
	}

	// Priority: (A) at the start
	priorityRe := regexp.MustCompile(`^\(([A-F])\)\s+`)
	if matches := priorityRe.FindStringSubmatch(input); matches != nil {
		t.Priority = Priority(matches[1][0])
		input = input[len(matches[0]):]
	}

	// Dates: created and completed (YYYY-MM-DD)
	dateRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	dates := dateRe.FindAllString(input, -1)
	if len(dates) > 0 {
		t.CreatedDate = dates[0]
		input = strings.Replace(input, dates[0], "", 1)
	}
	if len(dates) > 1 {
		t.CompletionDate = dates[1]
		input = strings.Replace(input, dates[1], "", 1)
	}

	// Projects: +project
	for _, word := range strings.Fields(input) {
		if strings.HasPrefix(word, "+") {
			t.Projects = append(t.Projects, word[1:])
		}
	}

	// Contexts: @context
	for _, word := range strings.Fields(input) {
		if strings.HasPrefix(word, "@") {
			t.Contexts = append(t.Contexts, word[1:])
		}
	}

	// Tags: key:value
	t.Tags = make(map[string]string)
	for _, word := range strings.Fields(input) {
		if strings.Contains(word, ":") {
			parts := strings.SplitN(word, ":", 2)
			t.Tags[parts[0]] = parts[1]
		}
	}

	words := []string{}
	for _, word := range strings.Fields(input) {
		if strings.HasPrefix(word, "+") || strings.HasPrefix(word, "@") || strings.Contains(word, ":") {
			break
		}
		words = append(words, word)
	}
	t.Name = strings.Join(words, " ")

	return t
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
