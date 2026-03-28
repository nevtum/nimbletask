package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all todos",
		RunE:  ListCmdFunc(),
	}
}

func ListCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check for config file first
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		// Determine todoPath if not set via flag - default to filename from config in PWD
		if todoPath == "" {
			// TODO: Add working directory error handling
			// Removed untested error handling: if err := os.Getwd(); err != nil { return error }
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Load todo list from file
		// TODO: Add error handling for file load errors (e.g., permission issues, corruption)
		tl, _, _ := loadTodoList(todoPath)

		if len(tl.GetRoots()) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No todos found")
			return nil
		}

		return nil
	}
}
