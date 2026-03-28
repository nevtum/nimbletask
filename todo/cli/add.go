package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// TODO: Add integration tests for file system edge cases (permissions, concurrent access)

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
		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		// Determine todoPath if not set via flag - default to filename from config in PWD
		// TODO: Add test coverage for os.Getwd error handling (rare edge case)
		if todoPath == "" {
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Load todo list from file
		// TODO: Add error handling for file load errors (e.g., permission issues, corruption)
		tl, _, _ := loadTodoList(todoPath)

		// Extract title from args
		title := args[0]

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
		// TODO: Add test coverage for Update error handling (e.g., concurrent modification)
		if finalPriority > 0 {
			_, _ = tl.Update(todoItem.ID, todo.TodoUpdate{Priority: &finalPriority})
		}

		if err := tl.Save(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Todo created! ID is %s\n", todoItem.ID)
		return nil
	}
}
