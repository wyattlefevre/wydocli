package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all configuration for wydoCLI.
// Priority order: CLI flags > config file > env vars > defaults
type Config struct {
	TodoDir  string `json:"todo_dir,omitempty"`
	TodoFile string `json:"todo_file,omitempty"`
	DoneFile string `json:"done_file,omitempty"`
	ProjDir  string `json:"proj_dir,omitempty"`
}

// CLIFlags holds command-line flag values that override other config sources
type CLIFlags struct {
	TodoDir string
}

var (
	// globalConfig is the loaded configuration
	globalConfig *Config

	// cliFlags holds CLI overrides
	cliFlags CLIFlags
)

// SetCLIFlags sets the CLI flag overrides (call before Load)
func SetCLIFlags(flags CLIFlags) {
	cliFlags = flags
}

// Load loads configuration with priority: CLI flags > config file > env vars > defaults
func Load() (*Config, error) {
	cfg := &Config{}

	// Start with defaults
	cfg.applyDefaults()

	// Layer 1: Apply env vars
	cfg.applyEnvVars()

	// Layer 2: Apply config file (if exists)
	if err := cfg.applyConfigFile(); err != nil {
		return nil, err
	}

	// Layer 3: Apply CLI flags (highest priority)
	cfg.applyCLIFlags()

	// Resolve relative paths
	cfg.resolvePaths()

	globalConfig = cfg
	return cfg, nil
}

// Get returns the loaded config, loading it if necessary
func Get() *Config {
	if globalConfig == nil {
		cfg, err := Load()
		if err != nil {
			// Fall back to defaults on error
			cfg = &Config{}
			cfg.applyDefaults()
			cfg.resolvePaths()
			globalConfig = cfg
		}
	}
	return globalConfig
}

// Reset clears the cached config (useful for testing)
func Reset() {
	globalConfig = nil
	cliFlags = CLIFlags{}
}

func (c *Config) applyDefaults() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	c.TodoDir = home
	c.TodoFile = "todo.txt"
	c.DoneFile = "done.txt"
	c.ProjDir = "todo_projects"
}

func (c *Config) applyEnvVars() {
	if val := os.Getenv("TODO_DIR"); val != "" {
		c.TodoDir = val
	}
	if val := os.Getenv("TODO_FILE"); val != "" {
		c.TodoFile = val
	}
	if val := os.Getenv("DONE_FILE"); val != "" {
		c.DoneFile = val
	}
	if val := os.Getenv("TODO_PROJ_DIR"); val != "" {
		c.ProjDir = val
	}
}

func (c *Config) applyConfigFile() error {
	configPath := getConfigPath()
	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Config file doesn't exist, that's fine
		}
		return err
	}

	// Parse JSON into a temporary struct to overlay non-empty values
	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return err
	}

	// Only override if values are set in the file
	if fileCfg.TodoDir != "" {
		c.TodoDir = fileCfg.TodoDir
	}
	if fileCfg.TodoFile != "" {
		c.TodoFile = fileCfg.TodoFile
	}
	if fileCfg.DoneFile != "" {
		c.DoneFile = fileCfg.DoneFile
	}
	if fileCfg.ProjDir != "" {
		c.ProjDir = fileCfg.ProjDir
	}

	return nil
}

func (c *Config) applyCLIFlags() {
	if cliFlags.TodoDir != "" {
		c.TodoDir = cliFlags.TodoDir
	}
}

func (c *Config) resolvePaths() {
	// Expand ~ in TodoDir
	c.TodoDir = expandPath(c.TodoDir)

	// If TodoFile/DoneFile/ProjDir are relative, make them relative to TodoDir
	if !filepath.IsAbs(c.TodoFile) {
		c.TodoFile = filepath.Join(c.TodoDir, c.TodoFile)
	}
	if !filepath.IsAbs(c.DoneFile) {
		c.DoneFile = filepath.Join(c.TodoDir, c.DoneFile)
	}
	if !filepath.IsAbs(c.ProjDir) {
		c.ProjDir = filepath.Join(c.TodoDir, c.ProjDir)
	}
}

// getConfigPath returns the path to the config file, or empty if not found
func getConfigPath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "wydo", "config.json")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fall back to ~/.config/wydo/config.json
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(home, ".config", "wydo", "config.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// expandPath expands ~ to the home directory
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}

	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}

	return path
}

// GetTodoDir returns the configured todo directory
func (c *Config) GetTodoDir() string {
	return c.TodoDir
}

// GetTodoFile returns the full path to todo.txt
func (c *Config) GetTodoFile() string {
	return c.TodoFile
}

// GetDoneFile returns the full path to done.txt
func (c *Config) GetDoneFile() string {
	return c.DoneFile
}

// GetProjDir returns the full path to the projects directory
func (c *Config) GetProjDir() string {
	return c.ProjDir
}
