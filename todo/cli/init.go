package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-config",
		Short: "Initialize the global configuration file",
		RunE:  InitCmdFunc(),
	}
}

// InitCmdFunc returns a RunE function for the init-config command
func InitCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Create parent directories if they don't exist
		if err := os.MkdirAll(configRoot, 0755); err != nil {
			return err
		}

		bytes, _ := json.Marshal(defaultConfig())
		configPath := filepath.Join(configRoot, "config.json")
		return os.WriteFile(configPath, bytes, 0644)
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
