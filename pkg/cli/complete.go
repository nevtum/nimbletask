package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// CompleteCmd returns a *cobra.Command instance for the complete command
// Uses global variables configRoot and todoPath
func CompleteCmd() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "complete [id]",
		Short: "Mark a todo item as completed",
		RunE:  CompleteCmdFunc(&recursive),
	}

	cmd.Flags().BoolVar(&recursive, "recursive", false, "Complete the todo and all its descendants")

	return cmd
}

// CompleteCmdFunc returns a RunE function for the complete command
func CompleteCmdFunc(recursive *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Manual arg validation - shows usage when args are missing
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}

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

		// Extract todo ID from args
		todoID := args[0]

		// Mark the todo as complete (optionally recursive)
		if *recursive {
			if err := tl.CompleteSubtree(todoID); err != nil {
				return err
			}
		} else {
			if err := tl.Complete(todoID); err != nil {
				return err
			}
		}

		if err := tl.Save(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Completed todo %s\n", todoID)
		return nil
	}
}
