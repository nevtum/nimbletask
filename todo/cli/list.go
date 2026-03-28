package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"todo_cli/todo"

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
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Load todo list from file
		tl, _, _ := loadTodoList(todoPath)

		roots := tl.GetRoots()
		if len(roots) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No todos found")
			return nil
		}

		// Display todos in numbered format
		displayTodos(cmd.OutOrStdout(), roots, "")

		return nil
	}
}

// displayTodos recursively displays todos with numbered paths
func displayTodos(out io.Writer, todos []*todo.Todo, prefix string) {
	for i, t := range todos {
		// Build the number for this todo
		var number string
		if prefix == "" {
			number = fmt.Sprintf("%d.", i+1)
		} else {
			number = fmt.Sprintf("%s.%d", prefix, i+1)
		}

		// Build checkbox
		checkbox := "[ ]"
		if t.Completed {
			checkbox = "[x]"
		}

		// Format: "1. [ ] ID Title"
		line := fmt.Sprintf("%s %s %s %s\n", number, checkbox, t.ID, t.Title)
		out.Write([]byte(line))

		// Recursively display children with updated prefix
		if len(t.Children) > 0 {
			displayTodos(out, t.Children, number)
		}
	}
}
