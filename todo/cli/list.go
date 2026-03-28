package cli

import (
	"fmt"
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
			// TODO: Add working directory error handling
			// Removed untested error handling: if err := os.Getwd(); err != nil { return error }
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		// Create a new todo list and load from file
		tl := todo.NewTodoList()

		file := todo.NewFile(todoPath)
		// TODO: Add todo list load error handling
		// Removed untested error handling: if err := tl.Load(file); err != nil { return error }
		_ = tl.Load(file)

		if len(tl.GetRoots()) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No todos found")
			return nil
		}

		return nil
	}
}
