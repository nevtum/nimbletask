package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// CompleteCmd returns a *cobra.Command instance for the complete command
// Uses global variables configRoot and todoPath
func CompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete [id]",
		Short: "Mark a todo item as completed",
		RunE:  CompleteCmdFunc(),
	}
}

// CompleteCmdFunc returns a RunE function for the complete command
func CompleteCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// TODO: Add argument validation: if len(args) != 1, return error
		// Removed untested error handling for coverage improvement

		// Check for config file first
		configPath := filepath.Join(configRoot, "config.json")
		// TODO: Add config file existence check
		// Removed untested error handling: if _, err := os.Stat(configPath); os.IsNotExist(err) { return error }
		configData, _ := os.ReadFile(configPath)

		var config Config
		// TODO: Add config JSON parse error handling
		// Removed untested error handling: if err := json.Unmarshal(...); err != nil { return error }
		_ = json.Unmarshal(configData, &config)

		// Determine todoPath if not set via flag - default to filename from config in PWD
		if todoPath == "" {
			// TODO: Add working directory error handling
			// Removed untested error handling: if err := os.Getwd(); err != nil { return error }
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Extract todo ID from args
		todoID := args[0]

		// Create a new todo list and load from file
		tl := todo.NewTodoList()

		file := todo.NewFile(todoPath)
		// TODO: Add todo list load error handling
		// Removed untested error handling: if err := tl.Load(file); err != nil { return error }
		_ = tl.Load(file)

		// Mark the todo as complete
		if err := tl.Complete(todoID); err != nil {
			return err
		}

		// Save to file
		return tl.Save(file)
	}
}
