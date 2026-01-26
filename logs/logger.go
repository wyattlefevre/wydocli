package logs

import (
    "log"
    "os"
    "path/filepath"
    "sync"
)

var (
    Logger  *log.Logger
    logFile *os.File
    mu      sync.Mutex
)

// This runs automatically when the package is imported.
// Creates a logger in the current directory as a fallback.
func init() {
    f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("failed to open debug file: %v", err)
    }
    logFile = f
    Logger = log.New(f, "[wydocli] ", log.LstdFlags|log.Lshortfile)
}

// Initialize reinitializes the logger to write to a new directory.
// This should be called after configuration is loaded to move the log file
// to the configured TODO_DIR. Returns an error if the new log file cannot
// be opened, but the logger will continue using the old location.
func Initialize(logDir string) error {
    mu.Lock()
    defer mu.Unlock()

    // Skip reinitialization if logDir is empty or current directory
    if logDir == "" || logDir == "." {
        return nil
    }

    logPath := filepath.Join(logDir, "debug.log")

    // Log the migration before switching
    Logger.Printf("Reinitializing logger to: %s", logPath)

    // Open new log file
    f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        Logger.Printf("Failed to open new log file at %s: %v", logPath, err)
        return err
    }

    // Close old log file
    if logFile != nil {
        logFile.Close()
    }

    // Update logger and file handle
    logFile = f
    Logger = log.New(f, "[wydocli] ", log.LstdFlags|log.Lshortfile)

    // Log success to new location
    Logger.Printf("Logger successfully reinitialized to: %s", logPath)

    return nil
}

// Close closes the log file. Useful for cleanup in tests.
func Close() error {
    mu.Lock()
    defer mu.Unlock()

    if logFile != nil {
        return logFile.Close()
    }
    return nil
}

