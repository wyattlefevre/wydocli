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
)

var (
	todoDir = getEnv("TODO_DIR", defaultTodoDir())
	projDir = getEnv("TODO_PROJ_DIR", filepath.Join(defaultTodoDir(), "todo_projects"))

	todoFilePath = getEnv("TODO_FILE", filepath.Join(todoDir, "todo.txt"))
	doneFilePath = getEnv("DONE_FILE", filepath.Join(todoDir, "done.txt"))

	mu sync.RWMutex

	todoFileTaskList []Task
	doneFileTaskList []Task

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

func LoadAllTasks() error {
	var err error
	projectMap = make(map[string]Project)

	err = scanProjectFiles(projectMap)
	if err != nil {
		return err
	}

	todoFileTaskList, err = loadTasks(todoFilePath, false, projectMap)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", todoFilePath, err)
	}
	doneFileTaskList, err = loadTasks(doneFilePath, false, projectMap)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", doneFilePath, err)
	}
	return nil
}

func CleanTasks() error {
	var err error
	projectMap = make(map[string]Project)
	todoFileTaskList, err = loadTasks(todoFilePath, true, projectMap)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", todoFilePath, err)
	}
	doneFileTaskList, err = loadTasks(doneFilePath, true, projectMap)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", doneFilePath, err)
	}
	writeTasks()
	return nil
}

func PrintTasks() {
	fmt.Println("---------------")
	fmt.Printf("TODO file Tasks: %d\n", len(todoFileTaskList))
	fmt.Println("---------------")
	for _, task := range todoFileTaskList {
		task.Print()
		fmt.Println("---")
	}
	fmt.Println("---------------")
	fmt.Printf("DONE file Tasks: %d\n", len(doneFileTaskList))
	fmt.Println("---------------")
	for _, task := range doneFileTaskList {
		task.Print()
		fmt.Println("---")
	}
}

func PrintProjects() {
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
		todo, done := TaskCount(project.Name)
		fmt.Printf("TODO: %d | DONE: %d\n", todo, done)
	}
	fmt.Println("---------------")
}

func TaskCount(project string) (int, int) {
	todoCount := 0
	doneCount := 0
	for _, task := range todoFileTaskList {
		if task.HasProject(project) {
			if task.Done {
				doneCount++
			} else {
				todoCount++
			}
		}
	}
	for _, task := range todoFileTaskList {
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

func writeTasks() error {
	mu.Lock()
	defer mu.Unlock()

	// Write todo tasks
	todoFile, err := os.Create(todoFilePath)
	if err != nil {
		return fmt.Errorf("Error writing %s: %v", todoFilePath, err)
	}
	defer todoFile.Close()
	for _, task := range todoFileTaskList {
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
	for _, task := range doneFileTaskList {
		_, err := fmt.Fprintln(doneFile, task.String())
		if err != nil {
			return fmt.Errorf("Error writing to %s: %v", doneFilePath, err)
		}
	}

	return nil
}

func loadTasks(filePath string, allowMismatch bool, projects map[string]Project) ([]Task, error) {
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
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // skip blank lines
		}
		hashId := HashTaskLine(line)
		task := ParseTask(line, hashId)
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
