package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Reset any previous state
	Reset()

	// Clear env vars that might interfere
	os.Unsetenv("TODO_DIR")
	os.Unsetenv("TODO_FILE")
	os.Unsetenv("DONE_FILE")
	os.Unsetenv("TODO_PROJ_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	home, _ := os.UserHomeDir()

	if cfg.TodoDir != home {
		t.Errorf("TodoDir = %q, want %q", cfg.TodoDir, home)
	}
	if cfg.TodoFile != filepath.Join(home, "todo.txt") {
		t.Errorf("TodoFile = %q, want %q", cfg.TodoFile, filepath.Join(home, "todo.txt"))
	}
	if cfg.DoneFile != filepath.Join(home, "done.txt") {
		t.Errorf("DoneFile = %q, want %q", cfg.DoneFile, filepath.Join(home, "done.txt"))
	}
}

func TestLoad_EnvVars(t *testing.T) {
	Reset()

	// Set env vars
	tmpDir := t.TempDir()
	os.Setenv("TODO_DIR", tmpDir)
	defer os.Unsetenv("TODO_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.TodoDir != tmpDir {
		t.Errorf("TodoDir = %q, want %q", cfg.TodoDir, tmpDir)
	}
	if cfg.TodoFile != filepath.Join(tmpDir, "todo.txt") {
		t.Errorf("TodoFile = %q, want %q", cfg.TodoFile, filepath.Join(tmpDir, "todo.txt"))
	}
}

func TestLoad_CLIFlags(t *testing.T) {
	Reset()

	// Clear env vars
	os.Unsetenv("TODO_DIR")

	// Set CLI flags
	tmpDir := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: tmpDir})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.TodoDir != tmpDir {
		t.Errorf("TodoDir = %q, want %q", cfg.TodoDir, tmpDir)
	}
}

func TestLoad_CLIFlagsOverrideEnvVars(t *testing.T) {
	Reset()

	// Set env var
	envDir := t.TempDir()
	os.Setenv("TODO_DIR", envDir)
	defer os.Unsetenv("TODO_DIR")

	// Set CLI flags (should take precedence)
	cliDir := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: cliDir})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.TodoDir != cliDir {
		t.Errorf("TodoDir = %q, want %q (CLI flags should override env vars)", cfg.TodoDir, cliDir)
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	Reset()

	// Clear env vars
	os.Unsetenv("TODO_DIR")
	os.Unsetenv("XDG_CONFIG_HOME")

	// Create a temp config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "wydo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create config file
	todoDir := filepath.Join(tmpDir, "my-todos")
	configContent := `{"todo_dir": "` + todoDir + `"}`
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Point XDG_CONFIG_HOME to our temp config
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.TodoDir != todoDir {
		t.Errorf("TodoDir = %q, want %q", cfg.TodoDir, todoDir)
	}
}

func TestGet_CachesConfig(t *testing.T) {
	Reset()

	tmpDir := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: tmpDir})

	cfg1 := Get()
	cfg2 := Get()

	if cfg1 != cfg2 {
		t.Error("Get() should return cached config")
	}
}

func TestReset_ClearsCache(t *testing.T) {
	Reset()

	tmpDir1 := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: tmpDir1})
	cfg1 := Get()

	Reset()

	tmpDir2 := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: tmpDir2})
	cfg2 := Get()

	if cfg1.TodoDir == cfg2.TodoDir {
		t.Error("Reset() should clear cached config")
	}
}

func TestExpandPath_Tilde(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/foo", filepath.Join(home, "foo")},
		{"~/", filepath.Join(home, "")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tc := range tests {
		result := expandPath(tc.input)
		if result != tc.expected {
			t.Errorf("expandPath(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestConfig_Getters(t *testing.T) {
	Reset()

	tmpDir := t.TempDir()
	SetCLIFlags(CLIFlags{TodoDir: tmpDir})

	cfg, _ := Load()

	if cfg.GetTodoDir() != tmpDir {
		t.Errorf("GetTodoDir() = %q, want %q", cfg.GetTodoDir(), tmpDir)
	}
	if cfg.GetTodoFile() != filepath.Join(tmpDir, "todo.txt") {
		t.Errorf("GetTodoFile() = %q, want %q", cfg.GetTodoFile(), filepath.Join(tmpDir, "todo.txt"))
	}
	if cfg.GetDoneFile() != filepath.Join(tmpDir, "done.txt") {
		t.Errorf("GetDoneFile() = %q, want %q", cfg.GetDoneFile(), filepath.Join(tmpDir, "done.txt"))
	}
	if cfg.GetProjDir() != filepath.Join(tmpDir, "todo_projects") {
		t.Errorf("GetProjDir() = %q, want %q", cfg.GetProjDir(), filepath.Join(tmpDir, "todo_projects"))
	}
}
