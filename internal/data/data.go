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
	todoDir      = getEnv("TODO_DIR", defaultTodoDir())
	todoFilePath = getEnv("TODO_FILE", filepath.Join(todoDir, "todo.txt"))
	doneFilePath = getEnv("DONE_FILE", filepath.Join(todoDir, "done.txt"))

	mu sync.RWMutex

	todoTaskList []*Task
	todoTaskMap  map[string]*Task
	doneTaskList []*Task
	doneTaskMap  map[string]*Task
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
	todoTaskList, todoTaskMap, err = loadTasks(todoFilePath, false)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", todoFilePath, err)
	}
	doneTaskList, doneTaskMap, err = loadTasks(doneFilePath, false)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", doneFilePath, err)
	}
	return nil
}

func MalformedDiff() {
	// TODO show malformed tasks next to the formatted version (used before cleaning)
}

func CleanTasks() error {
	var err error
	todoTaskList, todoTaskMap, err = loadTasks(todoFilePath, true)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", todoFilePath, err)
	}
	doneTaskList, doneTaskMap, err = loadTasks(doneFilePath, true)
	if err != nil {
		return fmt.Errorf("Error reading %s: %v", doneFilePath, err)
	}
	writeTasks()
	return nil
}

func PrintTasks() {
	fmt.Println("---------------")
	fmt.Printf("TODO file Tasks: %d\n", len(todoTaskList))
	fmt.Println("---------------")
	for _, task := range todoTaskList {
		task.Print()
		fmt.Println("---")
	}
	fmt.Println("---------------")
	fmt.Printf("DONE file Tasks: %d\n", len(doneTaskList))
	fmt.Println("---------------")
	for _, task := range doneTaskList {
		task.Print()
		fmt.Println("---")
	}
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
	for _, task := range todoTaskList {
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
	for _, task := range doneTaskList {
		_, err := fmt.Fprintln(doneFile, task.String())
		if err != nil {
			return fmt.Errorf("Error writing to %s: %v", doneFilePath, err)
		}
	}

	return nil
}

func loadTasks(filePath string, allowMismatch bool) ([]*Task, map[string]*Task, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	taskList := []*Task{}
	taskMap := make(map[string]*Task)

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // skip blank lines
		}
		hashId := HashTaskLine(line)
		task := ParseTask(line, hashId)
		if task.String() != line && !allowMismatch {
			return nil, nil, &ParseTaskMismatchError{Msg: "Malformatted task detected in todo file"}
		}
		taskMap[hashId] = &task
		taskList = append(taskList, &task)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return taskList, taskMap, nil
}
