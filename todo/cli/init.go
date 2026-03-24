package cli

import (
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

		configPath := filepath.Join(configRoot, "config.json")
		return os.WriteFile(configPath, []byte("{}"), 0644)
	}
}
