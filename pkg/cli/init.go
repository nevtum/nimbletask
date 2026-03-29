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
		Use:   "init",
		Short: "Initialize the global configuration file",
		RunE:  InitCmdFunc(),
	}
}

// InitCmdFunc returns a RunE function for the init command
func InitCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Create parent directories if they don't exist
		if err := os.MkdirAll(configRoot, 0755); err != nil {
			return err
		}

		configPath := filepath.Join(configRoot, "config.json")

		// Check if config file already exists - don't overwrite
		if _, err := os.Stat(configPath); err == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Config file already exists at %s\n", configPath)
			return nil
		}

		bytes, _ := json.Marshal(defaultConfig())
		err := os.WriteFile(configPath, bytes, 0644)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Config file created at %s\n", configPath)
		return nil
	}
}
