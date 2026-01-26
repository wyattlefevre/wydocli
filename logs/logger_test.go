package logs

import (
    "os"
    "path/filepath"
    "testing"
)

func TestInitialize(t *testing.T) {
    // Create a temporary directory for testing
    tmpDir, err := os.MkdirTemp("", "wydo-log-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    // Test reinitialization
    err = Initialize(tmpDir)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }

    // Verify the log file was created
    logPath := filepath.Join(tmpDir, "debug.log")
    if _, err := os.Stat(logPath); os.IsNotExist(err) {
        t.Errorf("Log file was not created at %s", logPath)
    }

    // Write a test message
    Logger.Println("Test message after reinitialization")

    // Verify the file has content
    content, err := os.ReadFile(logPath)
    if err != nil {
        t.Fatalf("Failed to read log file: %v", err)
    }

    if len(content) == 0 {
        t.Error("Log file is empty after writing")
    }
}

func TestInitialize_EmptyDir(t *testing.T) {
    // Test that empty directory is skipped
    err := Initialize("")
    if err != nil {
        t.Errorf("Initialize with empty dir should not error, got: %v", err)
    }
}

func TestInitialize_CurrentDir(t *testing.T) {
    // Test that current directory is skipped
    err := Initialize(".")
    if err != nil {
        t.Errorf("Initialize with current dir should not error, got: %v", err)
    }
}

func TestInitialize_NonExistentDir(t *testing.T) {
    // Test with non-existent directory
    err := Initialize("/this/directory/does/not/exist/hopefully")
    if err == nil {
        t.Error("Initialize with non-existent dir should return error")
    }
}
