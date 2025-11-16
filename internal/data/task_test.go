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
			name:  "Tag with space after colon (should not be detected)",
			input: "(A) Buy milk cost: 1000",
			expected: Task{
				Priority: PriorityA,
				Name:     "Buy milk cost: 1000",
			},
		},
		{
			name:  "Tag with space before colon (should not be detected)",
			input: "(A) Buy milk cost :1000",
			expected: Task{
				Priority: PriorityA,
				Name:     "Buy milk cost :1000",
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
			got := ParseTask(tc.input, "abc", "file.txt")
			if !tasksEqual(got, tc.expected) {
				t.Errorf("Test '%s' failed.\n%s", tc.name, diffTasks(tc.expected, got))
			}
		})
	}
}

func TestTask_String(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Priority and Name",
			input:    "(A) Buy milk",
			expected: "(A) Buy milk",
		},
		{
			name:     "Name Only",
			input:    "Buy milk",
			expected: "Buy milk",
		},
		{
			name:     "Completed with Dates",
			input:    "x (B) 2023-01-01 2023-01-02 Finish report",
			expected: "x (B) 2023-01-01 2023-01-02 Finish report",
		},
		{
			name:     "Projects, Contexts, Tags",
			input:    "(C) Plan trip +vacation @home cost:1000",
			expected: "(C) Plan trip +vacation @home cost:1000",
		},
		{
			name:     "Multiple Projects and Contexts",
			input:    "(A) Plan trip +vacation +workshop @home @office cost:1000",
			expected: "(A) Plan trip +vacation +workshop @home @office cost:1000",
		},
		{
			name:     "Incorrectly Formatted Task (fields out of order)",
			input:    "+vacation @home cost:1000 (B) Plan trip",
			expected: "+vacation @home cost:1000",
		},
		{
			name:     "Incorrectly Placed Priority",
			input:    "x 2023-01-01 2023-01-02 (B) Finish report",
			expected: "x 2023-01-01 2023-01-02 (B) Finish report",
		},
		{
			name:     "handles colons with trailing space",
			input:    "Buy milk cost: 1000",
			expected: "Buy milk cost: 1000",
		},
		{
			name:     "handles colons with prefixed space",
			input:    "Buy milk cost :1000",
			expected: "Buy milk cost :1000",
		},
		{
			name:     "trims whitespace",
			input:    "  Buy  milk ",
			expected: "Buy milk",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := ParseTask(tc.input, "abc", "file.txt")
			got := task.String()
			if got != tc.expected {
				t.Errorf("Test '%s' failed.\nExpected: %q\nGot:      %q", tc.name, tc.expected, got)
			}
		})
	}
}

func TestTask_String_FromStruct(t *testing.T) {
	cases := []struct {
		name     string
		task     Task
		expected string
	}{
		{
			name:     "Priority and Name",
			task:     Task{Priority: PriorityA, Name: "Buy milk"},
			expected: "(A) Buy milk",
		},
		{
			name:     "Completed with Dates",
			task:     Task{Done: true, Priority: PriorityB, CreatedDate: "2023-01-01", CompletionDate: "2023-01-02", Name: "Finish report"},
			expected: "x (B) 2023-01-01 2023-01-02 Finish report",
		},
		{
			name:     "Projects, Contexts, Tags",
			task:     Task{Priority: PriorityC, Name: "Plan trip", Projects: []string{"vacation"}, Contexts: []string{"home"}, Tags: map[string]string{"cost": "1000"}},
			expected: "(C) Plan trip +vacation @home cost:1000",
		},
		{
			name:     "Multiple Projects and Contexts",
			task:     Task{Priority: PriorityA, Name: "Plan trip", Projects: []string{"vacation", "workshop"}, Contexts: []string{"home", "office"}, Tags: map[string]string{"cost": "1000"}},
			expected: "(A) Plan trip +vacation +workshop @home @office cost:1000",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.task.String()
			if got != tc.expected {
				t.Errorf("Test '%s' failed.\nExpected: %q\nGot:      %q", tc.name, tc.expected, got)
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

func TestFirstProjectIndex_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"no tag", "abc", -1},
		{"tag at start", "+a bc", -1},
		{"tag after space", "ab +c", 3},
		{"tag after tab", "ab\t+c", 3},
		{"multiple tags", "ab +c +d", 3},
		{"tag with number", "ab +1", 3},
		{"tag with letter", "ab +a", 3},
		{"tag not preceded by space", "ab+c +a", 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstProjectIndex(tc.input)
			if got != tc.expected {
				t.Errorf("Test '%s' failed. Expected: %d, Got: %d", tc.name, tc.expected, got)
			}
		})
	}
}

func TestFirstContextIndex_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"no tag", "abc", -1},
		{"tag at start", "@a bc", -1},
		{"tag after space", "ab @c", 3},
		{"tag after tab", "ab\t@c", 3},
		{"multiple tags", "ab @c @d", 3},
		{"tag with number", "ab @1", 3},
		{"tag with letter", "ab @a", 3},
		{"tag not preceded by space", "ab@c @a", 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstContextIndex(tc.input)
			if got != tc.expected {
				t.Errorf("Test '%s' failed. Expected: %d, Got: %d", tc.name, tc.expected, got)
			}
		})
	}
}

func TestFirstTagIndex_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"no tag", "abc", -1},
		{"tag at start", "a:bc", -1},
		{"tag at start after space", " a:bc", 1},
		{"tag after space", "ab cost:1000", 3},
		{"multiple tags", "ab cost:1000 foo:bar", 3},
		{"tag with number", "ab 1:2", 3},
		{"tag with letter", "ab a:b", 3},
		{"tag with space before colon", "ab cost :1000", -1},
		{"tag with space after colon", "ab cost: 1000", -1},
		{"tag with non-alphanumeric", "ab cost-1:1000", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstTagIndex(tc.input)
			if got != tc.expected {
				t.Errorf("Test '%s' failed. Expected: %d, Got: %d", tc.name, tc.expected, got)
			}
		})
	}
}

func TestParsePriority_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Priority
	}{
		{"no priority", "Buy milk", PriorityNone},
		{"priority A uppercase", "(A) Buy milk", PriorityA},
		{"priority B uppercase", "(B) Buy milk", PriorityB},
		{"priority C uppercase", "(C) Buy milk", PriorityC},
		{"priority D uppercase", "(D) Buy milk", PriorityD},
		{"priority E uppercase", "(E) Buy milk", PriorityE},
		{"priority F uppercase", "(F) Buy milk", PriorityF},
		{"priority a lowercase", "(a) Buy milk", PriorityA},
		{"priority b lowercase", "(b) Buy milk", PriorityB},
		{"priority f lowercase", "(f) Buy milk", PriorityF},
		{"priority not at start", "Buy milk (A)", PriorityNone},
		{"invalid priority", "(G) Buy milk", PriorityNone},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParsePriority(tc.input)
			if got != tc.expected {
				t.Errorf("Test '%s' failed. Expected: %v, Got: %v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestParseProjects_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"no project", "abc", []string{}},
		{"single project", "do +work", []string{"work"}},
		{"multiple projects", "do +work +play", []string{"work", "play"}},
		{"project at start", "+start abc", []string{}},
		{"project with number", "do +p1", []string{"p1"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseProjects(tc.input)
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("Test '%s' failed. Expected: %#v, Got: %#v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestParseContexts_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"no context", "abc", []string{}},
		{"single context", "do @home", []string{"home"}},
		{"multiple contexts", "do @home @office", []string{"home", "office"}},
		{"context at start", "@start abc", []string{"start"}},
		{"context with number", "do @c1", []string{"c1"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseContexts(tc.input)
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("Test '%s' failed. Expected: %#v, Got: %#v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestParseTags_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"no tag", "abc", []string{}},
		{"single tag", "do cost:1000", []string{"cost:1000"}},
		{"multiple tags", "do cost:1000 foo:bar", []string{"cost:1000", "foo:bar"}},
		{"tag with number", "do 1:2", []string{"1:2"}},
		{"tag with letter", "do a:b", []string{"a:b"}},
		{"tag with space before colon", "do cost :1000", []string{}},
		{"tag with space after colon", "do cost: 1000", []string{}},
		{"tag with non-alphanumeric", "do cost-1:1000", []string{}},
		{"tag preceded by tab", "do\tcost:1000", []string{"cost:1000"}},
		{"tag at beginning", "cost-1:1000", []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseTags(tc.input)
			if !equalStringSlices(got, tc.expected) {
				t.Errorf("Test '%s' failed. Expected: %#v, Got: %#v", tc.name, tc.expected, got)
			}
		})
	}
}

func TestFirstMetaIndex_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		i1, i2, i3 int
		expected int
	}{
		{"all -1", -1, -1, -1, -1},
		{"one positive", 2, -1, -1, 2},
		{"two positive, first smaller", 1, 3, -1, 1},
		{"two positive, second smaller", 5, 2, -1, 2},
		{"three positive, middle smallest", 7, 1, 3, 1},
		{"all positive, last smallest", 4, 2, 0, 0},
		{"all zero", 0, 0, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstMetaIndex(tc.i1, tc.i2, tc.i3)
			if got != tc.expected {
				t.Errorf("Test '%s' failed. Expected: %d, Got: %d", tc.name, tc.expected, got)
			}
		})
	}
}
