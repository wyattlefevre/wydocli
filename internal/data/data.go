package data

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wyattlefevre/wydocli/internal/config"
	"github.com/wyattlefevre/wydocli/logs"
)

var (
	mu sync.RWMutex

	projectMap map[string]Project
)

// Path accessor functions that use the config package
func getTodoFilePath() string {
	return config.Get().GetTodoFile()
}

func getDoneFilePath() string {
	return config.Get().GetDoneFile()
}

func getProjDir() string {
	return config.Get().GetProjDir()
}

func HashTaskLine(line string) string {
	h := sha1.New()
	h.Write([]byte(line))
	return hex.EncodeToString(h.Sum(nil))[:10] // shorten to 10 chars for readability
}

type ParseTaskMismatchError struct {
	Msg string
}

func (e *ParseTaskMismatchError) Error() string {
	return e.Msg
}

func UpdateTask(tasks []Task, updatedTask Task) []Task {
	logs.Logger.Printf("Update Task: %s\n", updatedTask)
	found := false
	for i, t := range tasks {
		if t.ID == updatedTask.ID {
			logs.Logger.Println("task found. updating...")
			tasks[i] = updatedTask
			found = true
			break
		}
	}
	if !found {
		logs.Logger.Println("task not found. adding new task...")
		tasks = append(tasks, updatedTask)
	}
	return tasks
}

func LoadData(allowMismatch bool) ([]Task, map[string]Project, error) {
	logs.Logger.Println("LoadData")
	var err error

	todoFilePath := getTodoFilePath()
	doneFilePath := getDoneFilePath()

	// Projects
	projectMap = make(map[string]Project)
	err = scanProjectFiles(projectMap)
	if err != nil {
		// Don't fail if project dir doesn't exist
		if !os.IsNotExist(err) {
			return nil, nil, err
		}
	}

	// Tasks
	logs.Logger.Println("load todo.txt")
	todoTasks, err := loadTaskFile(todoFilePath, allowMismatch, projectMap)
	if err != nil {
		if _, ok := err.(*ParseTaskMismatchError); ok {
			logs.Logger.Printf("ParseTaskMismatchError: %v\n", err)
			return nil, nil, err
		}
		// Don't fail if todo.txt doesn't exist
		if !os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("Error reading %s: %v", todoFilePath, err)
		}
		todoTasks = []Task{}
	}

	logs.Logger.Println("load done.txt")
	doneTasks, err := loadTaskFile(doneFilePath, allowMismatch, projectMap)
	if err != nil {
		// Don't fail if done.txt doesn't exist
		if !os.IsNotExist(err) {
			logs.Logger.Fatalf("Error reading file %v", err)
			return nil, nil, fmt.Errorf("Error reading %s: %v", doneFilePath, err)
		}
		doneTasks = []Task{}
	}

	allTasks := append(todoTasks, doneTasks...)
	return allTasks, projectMap, nil
}

func WriteData(tasks []Task) error {
	todoFilePath := getTodoFilePath()
	doneFilePath := getDoneFilePath()

	logs.Logger.Printf("WriteData (%d tasks)", len(tasks))
	mu.Lock()
	defer mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(todoFilePath), 0755); err != nil {
		return fmt.Errorf("Error creating directory: %v", err)
	}

	// Write todo tasks
	todoFile, err := os.Create(todoFilePath)
	if err != nil {
		return fmt.Errorf("Error writing %s: %v", todoFilePath, err)
	}
	defer todoFile.Close()
	for _, task := range tasks {
		if task.File != todoFilePath {
			continue
		}
		_, err := fmt.Fprintln(todoFile, task.String())
		if err != nil {
			return fmt.Errorf("Error writing to %s: %v", todoFilePath, err)
		}
	}

	// Write done tasks
	doneFile, err := os.Create(doneFilePath)
	if err != nil {
		return fmt.Errorf("Error writing %s: %v", doneFilePath, err)
	}
	defer doneFile.Close()
	for _, task := range tasks {
		if task.File != doneFilePath {
			continue
		}
		task.Done = true
		_, err := fmt.Fprintln(doneFile, task.String())
		if err != nil {
			return fmt.Errorf("Error writing to %s: %v", doneFilePath, err)
		}
	}

	return nil
}

func PrintTasks(tasks []Task) {
	fmt.Println("---------------")
	fmt.Printf("Tasks: %d\n", len(tasks))
	fmt.Println("---------------")
	for _, task := range tasks {
		task.Print()
		fmt.Println("---")
	}
}

func PrintProjects(tasks []Task) {
	fmt.Println("---------------")
	fmt.Printf("Projects: %d\n", len(projectMap))
	fmt.Println("---------------")
	for name, project := range projectMap {
		fmt.Printf("\nProject: %s\n", name)
		if project.NotePath != nil {
			fmt.Printf("NotePath: %s\n", *project.NotePath)
		} else {
			fmt.Printf("NotePath: nil\n")
		}
		todo, done := TaskCount(tasks, project.Name)
		fmt.Printf("TODO: %d | DONE: %d\n", todo, done)
	}
	fmt.Println("---------------")
}

func TaskCount(tasks []Task, project string) (int, int) {
	todoCount := 0
	doneCount := 0
	for _, task := range tasks {
		if task.HasProject(project) {
			if task.Done {
				doneCount++
			} else {
				todoCount++
			}
		}
	}
	return todoCount, doneCount
}

func ArchiveDone(tasks []Task) error {
	doneFilePath := getDoneFilePath()
	for i := range tasks {
		if tasks[i].Done {
			tasks[i].File = doneFilePath
		}
	}
	err := WriteData(tasks)
	return err
}

func scanProjectFiles(projectMap map[string]Project) error {
	projDir := getProjDir()
	return filepath.Walk(projDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		if name == "" {
			return nil
		}
		relPath, relErr := filepath.Rel(projDir, path)
		if relErr != nil {
			return relErr
		}
		if _, exists := projectMap[name]; !exists {
			projectMap[name] = Project{
				Name:     name,
				NotePath: &relPath,
			}
		} else {
			proj := projectMap[name]
			proj.NotePath = &relPath
			projectMap[name] = proj
		}
		return nil
	})
}

func loadTaskFile(filePath string, allowMismatch bool, projects map[string]Project) ([]Task, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	taskList := []Task{}

	// Read file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		if strings.TrimSpace(line) == "" {
			continue // skip blank lines
		}
		hashId := HashTaskLine(fmt.Sprintf("%d:%s", lineNum, filePath))
		task := ParseTask(line, hashId, filePath)
		for _, project := range task.Projects {
			if _, exists := projects[project]; !exists {
				projects[project] = Project{Name: project}
			}
		}
		if task.String() != line && !allowMismatch {
			msg := fmt.Sprintf("malformed task\nparsed: %s\noriginal: %s", task.String(), line)
			logs.Logger.Println(msg)
			return nil, &ParseTaskMismatchError{Msg: msg}
		}
		taskList = append(taskList, task)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return taskList, nil
}

// DeleteTask removes a task by ID from the task slice and returns the updated slice.
func DeleteTask(tasks []Task, id string) []Task {
	for i, t := range tasks {
		if t.ID == id {
			return append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}

// AppendTask appends a single task line to the todo.txt file efficiently.
// It parses the line, assigns an ID, and returns the created Task.
func AppendTask(rawLine string) (*Task, error) {
	todoFilePath := getTodoFilePath()

	mu.Lock()
	defer mu.Unlock()

	rawLine = strings.TrimSpace(rawLine)
	if rawLine == "" {
		return nil, fmt.Errorf("empty task line")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(todoFilePath), 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %v", err)
	}

	// Count existing lines to generate a unique ID
	lineCount := 0
	file, err := os.Open(todoFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error opening %s: %v", todoFilePath, err)
	}
	if file != nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) != "" {
				lineCount++
			}
		}
		file.Close()
	}

	// Generate ID for the new task
	hashId := HashTaskLine(fmt.Sprintf("%d:%s", lineCount+1, todoFilePath))
	task := ParseTask(rawLine, hashId, todoFilePath)

	// Append to file
	f, err := os.OpenFile(todoFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening %s for append: %v", todoFilePath, err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, task.String())
	if err != nil {
		return nil, fmt.Errorf("error writing to %s: %v", todoFilePath, err)
	}

	return &task, nil
}

// GetTodoFilePath returns the configured path to todo.txt
func GetTodoFilePath() string {
	return getTodoFilePath()
}

// GetDoneFilePath returns the configured path to done.txt
func GetDoneFilePath() string {
	return getDoneFilePath()
}
