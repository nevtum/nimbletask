package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// priorityFlag is the local flag variable for the --priority flag
var priorityFlag int

// AddCmd returns a *cobra.Command instance for the add command
// Uses global variables configRoot and todoPath
func AddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new todo item",
		RunE:  AddCmdFunc(),
	}

	// Add --priority flag: priority level for the new todo (overrides config default)
	cmd.Flags().IntVar(&priorityFlag, "priority", 0, "Priority level for the new todo (overrides config default)")

	return cmd
}

func AddCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Manual arg validation - shows usage when args are missing
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}

		// Check for config file first
		configPath := filepath.Join(configRoot, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Silence usage for config errors - only show error message
			cmd.SilenceUsage = true
			return fmt.Errorf("config file not found at %s: init-config must be called first", configPath)
		}

		// Read and parse config file
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		var config Config
		if err := json.Unmarshal(configData, &config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}

		// Determine todoPath if not set via flag - default to filename from config in PWD
		if todoPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Extract title from args
		title := args[0]

		// Create a new todo list
		tl := todo.NewTodoList()

		file := todo.NewFile(todoPath)
		if err := tl.Load(file); err != nil {
			return err
		}

		// Add the todo (no parent, append to end)
		todoItem, err := tl.Add(title, "", -1)
		if err != nil {
			return err
		}

		// Determine priority: flag value takes precedence over config default
		var finalPriority int
		if cmd.Flags().Changed("priority") {
			// User explicitly provided --priority flag
			finalPriority = priorityFlag
		} else if config.DefaultPriority > 0 {
			// Use config default when flag not provided
			finalPriority = config.DefaultPriority
		}

		// Apply priority if set (either from flag or config)
		if finalPriority > 0 {
			_, err = tl.Update(todoItem.ID, todo.TodoUpdate{Priority: &finalPriority})
			if err != nil {
				return err
			}
		}

		// Save to file
		return tl.Save(file)
	}
}
