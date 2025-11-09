package data

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseTask_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Task
	}{
		{
			name:  "Priority and Name",
			input: "(A) Buy milk",
			expected: Task{
				Priority: PriorityA,
				Name:     "Buy milk",
			},
		},
		{
			name:  "Completed with Dates",
			input: "x (B) 2023-01-01 2023-01-02 Finish report",
			expected: Task{
				Done:           true,
				Priority:       PriorityB,
				CreatedDate:    "2023-01-01",
				CompletionDate: "2023-01-02",
				Name:           "Finish report",
			},
		},
		{
			name:  "Projects, Contexts, Tags",
			input: "(C) Plan trip +vacation @home cost:1000",
			expected: Task{
				Priority: PriorityC,
				Name:     "Plan trip",
				Projects: []string{"vacation"},
				Contexts: []string{"home"},
				Tags:     map[string]string{"cost": "1000"},
			},
		},
		{
			name:  "Multiple Projects and Contexts",
			input: "(A) Plan trip +vacation +workshop @home @office cost:1000",
			expected: Task{
				Priority: PriorityA,
				Name:     "Plan trip",
				Projects: []string{"vacation", "workshop"},
				Contexts: []string{"home", "office"},
				Tags:     map[string]string{"cost": "1000"},
			},
		},
		{
			name:  "Incorrectly Formatted Task (fields out of order)",
			input: "+vacation @home cost:1000 (B) Plan trip",
			expected: Task{
				Priority: 0,
				Name:     "",
				Projects: []string{"vacation"},
				Contexts: []string{"home"},
				Tags:     map[string]string{"cost": "1000"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseTask(tc.input)
			if !tasksEqual(got, tc.expected) {
				t.Errorf("Test '%s' failed.\n%s", tc.name, diffTasks(tc.expected, got))
			}
		})
	}
}

// Helper function to compare two Task structs
func tasksEqual(a, b Task) bool {
	if a.Priority != b.Priority ||
		a.Name != b.Name ||
		a.Done != b.Done ||
		a.CreatedDate != b.CreatedDate ||
		a.CompletionDate != b.CompletionDate ||
		len(a.Projects) != len(b.Projects) ||
		len(a.Contexts) != len(b.Contexts) ||
		len(a.Tags) != len(b.Tags) {
		return false
	}
	for i := range a.Projects {
		if a.Projects[i] != b.Projects[i] {
			return false
		}
	}
	for i := range a.Contexts {
		if a.Contexts[i] != b.Contexts[i] {
			return false
		}
	}
	for k, v := range a.Tags {
		if b.Tags[k] != v {
			return false
		}
	}
	return true
}

// Pretty diff for mismatched fields
func diffTasks(expected, got Task) string {
	var out strings.Builder
	if expected.Priority != got.Priority {
		out.WriteString(
			fmt.Sprintf("Priority:\n  expected: %v\n  got:      %v\n", expected.Priority, got.Priority))
	}
	if expected.Name != got.Name {
		out.WriteString(
			fmt.Sprintf("Name:\n  expected: %q\n  got:      %q\n", expected.Name, got.Name))
	}
	if expected.Done != got.Done {
		out.WriteString(
			fmt.Sprintf("Done:\n  expected: %v\n  got:      %v\n", expected.Done, got.Done))
	}
	if expected.CreatedDate != got.CreatedDate {
		out.WriteString(
			fmt.Sprintf("CreatedDate:\n  expected: %q\n  got:      %q\n", expected.CreatedDate, got.CreatedDate))
	}
	if expected.CompletionDate != got.CompletionDate {
		out.WriteString(
			fmt.Sprintf("CompletionDate:\n  expected: %q\n  got:      %q\n", expected.CompletionDate, got.CompletionDate))
	}
	if !equalStringSlices(expected.Projects, got.Projects) {
		out.WriteString(
			fmt.Sprintf("Projects:\n  expected: %#v\n  got:      %#v\n", expected.Projects, got.Projects))
	}
	if !equalStringSlices(expected.Contexts, got.Contexts) {
		out.WriteString(
			fmt.Sprintf("Contexts:\n  expected: %#v\n  got:      %#v\n", expected.Contexts, got.Contexts))
	}
	if !equalStringMaps(expected.Tags, got.Tags) {
		out.WriteString(
			fmt.Sprintf("Tags:\n  expected: %#v\n  got:      %#v\n", expected.Tags, got.Tags))
	}
	return out.String()
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStringMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
