package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// InitCmd returns a RunE function for the init-config command
func InitCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		configPath := filepath.Join(configRoot, "config.json")
		return os.WriteFile(configPath, []byte("{}"), 0644)
	}
}
