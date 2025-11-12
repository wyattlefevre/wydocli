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

	"github.com/wyattlefevre/wydocli/logs"
)

var (
	todoDir = getEnv("TODO_DIR", defaultTodoDir())
	projDir = getEnv("TODO_PROJ_DIR", filepath.Join(defaultTodoDir(), "todo_projects"))

	todoFilePath = getEnv("TODO_FILE", filepath.Join(todoDir, "todo.txt"))
	doneFilePath = getEnv("DONE_FILE", filepath.Join(todoDir, "done.txt"))

	mu sync.RWMutex

	projectMap map[string]Project
)

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func defaultTodoDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// fallback if home can’t be determined
		return "."
	}
	return home
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

func UpdateTask(tasks []Task, updatedTask Task) {
	logs.Logger.Printf("Update Task: %s\n", updatedTask)
	for i, t := range tasks {
		if t.ID == updatedTask.ID {
			logs.Logger.Println("task found. updating...")
			tasks[i] = updatedTask
		}
	}
}

func LoadData(allowMismatch bool) ([]Task, map[string]Project, error) {
	logs.Logger.Println("LoadData")
	var err error

	// Projects
	projectMap = make(map[string]Project)
	err = scanProjectFiles(projectMap)
	if err != nil {
		return nil, nil, err
	}

	// Tasks
	todoTasks, err := loadTaskFile(todoFilePath, allowMismatch, projectMap)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading %s: %v", todoFilePath, err)
	}
	doneTasks, err := loadTaskFile(doneFilePath, allowMismatch, projectMap)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading %s: %v", doneFilePath, err)
	}
	allTasks := append(todoTasks, doneTasks...)
	return allTasks, projectMap, nil
}

func WriteData(tasks []Task) error {
	logs.Logger.Println("WriteData")
	mu.Lock()
	defer mu.Unlock()

	// Write todo tasks
	todoFile, err := os.Create(todoFilePath)
	if err != nil {
		return fmt.Errorf("Error writing %s: %v", todoFilePath, err)
	}
	defer todoFile.Close()
	for _, task := range tasks {
		logs.Logger.Printf("write '%s'", task.String())
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
	for _, task := range tasks {
		if task.Done {
			task.File = doneFilePath
		}
	}
	err := WriteData(tasks)
	return err
}

func scanProjectFiles(projectMap map[string]Project) error {
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
		hashId := HashTaskLine(fmt.Sprintf("%d:%s:%s", lineNum, filePath, line))
		task := ParseTask(line, hashId, filePath)
		for _, project := range task.Projects {
			if _, exists := projects[project]; !exists {
				projects[project] = Project{Name: project}
			}
		}
		if task.String() != line && !allowMismatch {
			return nil, &ParseTaskMismatchError{Msg: "Malformatted task detected in todo file"}
		}
		taskList = append(taskList, task)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return taskList, nil
}
