package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nevtum/nimbletask/pkg/todo"
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
		displayTodos(cmd.OutOrStdout(), tl.GetMinIDLength(), roots, "")

		return nil
	}
}

// displayTodos recursively displays todos with numbered paths
func displayTodos(out io.Writer, minIDLength int, todos []*todo.Todo, prefix string) {
	for i, t := range todos {
		// Build the number for this todo
		var number string
		if prefix == "" {
			number = fmt.Sprintf("%d.", i+1)
		} else {
			// Remove trailing dot from prefix to avoid double dots (e.g., "1..1" instead of "1.1")
			number = fmt.Sprintf("%s.%d", strings.TrimSuffix(prefix, "."), i+1)
		}

		// Build checkbox
		checkbox := "[ ]"
		if t.Completed {
			checkbox = "[x]"
		}

		// Use the minimum unique ID length for display
		displayID := t.ID
		if len(t.ID) > minIDLength {
			displayID = t.ID[:minIDLength]
		}

		// Format: "1. [ ] <id:ABC> Title"
		line := fmt.Sprintf("%s %s <id:%s> %s\n", number, checkbox, displayID, t.Title)
		out.Write([]byte(line))

		// Recursively display children with updated prefix
		if len(t.Children) > 0 {
			displayTodos(out, minIDLength, t.Children, number)
		}
	}
}
