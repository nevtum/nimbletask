package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Config represents the configuration structure from config.json
type Config struct {
	Filename        string `json:"filename"`
	DefaultPriority int    `json:"default_priority"`
}

func defaultConfig() Config {
	return Config{
		Filename:        "todos.md",
		DefaultPriority: 3,
	}
}

func loadConfig(cmd *cobra.Command) (*Config, error) {
	configPath := filepath.Join(configRoot, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cmd.SilenceUsage = true
		return nil, fmt.Errorf("config file not found at %s: init-config must be called first", configPath)
	}

	// Read and parse config file
	// TODO: Add test coverage for os.ReadFile error handling (e.g., permission changes after stat)
	configData, _ := os.ReadFile(configPath)

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &config, nil
}
