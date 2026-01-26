package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wyattlefevre/wydocli/internal/config"
	"github.com/wyattlefevre/wydocli/internal/service"
)

func setupTestService(t *testing.T, testdataDir string) service.TaskService {
	t.Helper()

	// Reset config for clean state
	config.Reset()

	// Get absolute path to testdata
	absPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", testdataDir))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Set CLI flags to use testdata directory
	config.SetCLIFlags(config.CLIFlags{TodoDir: absPath})

	// Load config
	_, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create service
	svc, err := service.NewTaskService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	return svc
}

func TestRunList_Basic(t *testing.T) {
	svc := setupTestService(t, "basic")

	// Test list returns success
	exitCode := runList([]string{}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_Empty(t *testing.T) {
	svc := setupTestService(t, "empty")

	exitCode := runList([]string{}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_Complex(t *testing.T) {
	svc := setupTestService(t, "complex")

	exitCode := runList([]string{}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_WithProjectFilter(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runList([]string{"-p", "work"}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_WithContextFilter(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runList([]string{"-c", "coding"}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_ShowDone(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runList([]string{"--done"}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunList_ShowAll(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runList([]string{"--all"}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRunAdd_RequiresDescription(t *testing.T) {
	svc := setupTestService(t, "empty")

	exitCode := runAdd([]string{}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing description, got %d", exitCode)
	}
}

func TestRunDone_RequiresID(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runDone([]string{}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing ID, got %d", exitCode)
	}
}

func TestRunDone_InvalidID(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runDone([]string{"nonexistent"}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid ID, got %d", exitCode)
	}
}

func TestRunDelete_RequiresID(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runDelete([]string{}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for missing ID, got %d", exitCode)
	}
}

func TestRunDelete_InvalidID(t *testing.T) {
	svc := setupTestService(t, "basic")

	exitCode := runDelete([]string{"nonexistent"}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for invalid ID, got %d", exitCode)
	}
}

func TestRun_Help(t *testing.T) {
	svc := setupTestService(t, "empty")

	exitCode := Run([]string{"help"}, svc)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	svc := setupTestService(t, "empty")

	exitCode := Run([]string{"unknown"}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for unknown command, got %d", exitCode)
	}
}

func TestRun_NoCommand(t *testing.T) {
	svc := setupTestService(t, "empty")

	exitCode := Run([]string{}, svc)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 for no command, got %d", exitCode)
	}
}

// TestAddDoneDeleteWorkflow tests the full lifecycle of a task using a temp directory
func TestAddDoneDeleteWorkflow(t *testing.T) {
	// Create a temp directory for this test
	tmpDir, err := os.MkdirTemp("", "wydo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Reset config and point to temp directory
	config.Reset()
	config.SetCLIFlags(config.CLIFlags{TodoDir: tmpDir})
	_, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	svc, err := service.NewTaskService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Step 1: Add a task
	exitCode := runAdd([]string{"Test workflow task", "+test"}, svc)
	if exitCode != 0 {
		t.Fatalf("Failed to add task, exit code: %d", exitCode)
	}

	// Step 2: Verify task exists
	tasks, err := svc.ListPending()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	taskID := tasks[0].ID

	// Step 3: Mark task as done
	exitCode = runDone([]string{taskID}, svc)
	if exitCode != 0 {
		t.Fatalf("Failed to complete task, exit code: %d", exitCode)
	}

	// Step 4: Verify task is done
	pendingTasks, _ := svc.ListPending()
	doneTasks, _ := svc.ListDone()
	if len(pendingTasks) != 0 {
		t.Errorf("Expected 0 pending tasks, got %d", len(pendingTasks))
	}
	if len(doneTasks) != 1 {
		t.Errorf("Expected 1 done task, got %d", len(doneTasks))
	}

	// Step 5: Delete the done task
	exitCode = runDelete([]string{doneTasks[0].ID}, svc)
	if exitCode != 0 {
		t.Fatalf("Failed to delete task, exit code: %d", exitCode)
	}

	// Step 6: Verify task is deleted
	allTasks, _ := svc.List()
	if len(allTasks) != 0 {
		t.Errorf("Expected 0 tasks after delete, got %d", len(allTasks))
	}
}
