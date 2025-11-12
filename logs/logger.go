package logs

import (
    "log"
    "os"
)

var Logger *log.Logger

// This runs automatically when the package is imported.
func init() {
    f, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("failed to open debug file: %v", err)
    }
    Logger = log.New(f, "[wydocli] ", log.LstdFlags|log.Lshortfile)
}

