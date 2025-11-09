package data

import (
	"strings"
	"regexp"
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
	Name           string
	Projects       []string
	Contexts       []string
	Done           bool
	Tags           map[string]string
	CreatedDate    string
	CompletionDate string
	Priority       Priority
}

func ParseTask(input string) Task {
	var t Task
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
