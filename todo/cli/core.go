package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// Global flag variables for CLI configuration
var (
	configRoot string
	todoPath   string
)

// NewRootCmd creates a new root command with all subcommands.
// This factory function enables testing by providing fresh command instances.
func NewRootCmd() *cobra.Command {

	// rootCmd is the main entry point for the CLI
	rootCmd := &cobra.Command{
		Use:   "todo",
		Short: "A CLI tool for managing hierarchical todo lists",
		Long:  `Manage your todo lists through a powerful CLI interface.`,
	}

	// Determine default config root (user's config directory)
	defaultConfigRoot := filepath.Join(os.Getenv("HOME"), ".config", "todos")

	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&configRoot, "config", defaultConfigRoot, "Configuration directory root")
	rootCmd.PersistentFlags().StringVar(&todoPath, "file", "", "Path to todo list file (default: todo.md in current directory)")

	// Add subcommands to root
	rootCmd.AddCommand(InitCmd())
	rootCmd.AddCommand(AddCmd())
	rootCmd.AddCommand(CompleteCmd())
	rootCmd.AddCommand(ListCmd())

	return rootCmd
}

// loadTodoList creates a new TodoList and loads it from the given file path.
// Returns a friendly error message if loading fails (e.g., permission issues, corruption).
func loadTodoList(path string) (*todo.TodoList, *todo.File, error) {
	tl := todo.NewTodoList()

	file := todo.NewFile(path)
	if err := tl.Load(file); err != nil {
		return nil, nil, fmt.Errorf("cannot load todos: %w", err)
	}

	return tl, file, nil
}
