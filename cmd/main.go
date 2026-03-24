package main

import (
	"fmt"
	"os"
	"path/filepath"

	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// main executes the root command
func main() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// NewRootCmd creates a new root command with all subcommands.
// This factory function enables testing by providing fresh command instances.
func NewRootCmd() *cobra.Command {
	// Local variables for this command instance
	var configRoot string
	var todoPath string

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

	// addCmd adds a new todo item to the list
	addCmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new todo item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine todoPath if not set via flag - default to todo.md in PWD
			if todoPath == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				todoPath = filepath.Join(cwd, "todo.md")
			}
			return runAdd(todoPath, args...)
		},
	}

	// initConfigCmd initializes the global configuration file
	initConfigCmd := &cobra.Command{
		Use:   "init-config",
		Short: "Initialize the global configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitConfig(configRoot)
		},
	}

	// Add subcommands to root
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(initConfigCmd)

	return rootCmd
}

// runInitConfig creates the configuration directory and file
func runInitConfig(configRoot string) error {
	configPath := filepath.Join(configRoot, "config.json")
	return os.WriteFile(configPath, []byte("{}"), 0644)
}

// runAdd creates a new todo and saves it to the todo list file
func runAdd(todoPath string, args ...string) error {
	// Extract title from args
	title := args[0]

	// Create a new todo list
	tl := todo.NewTodoList()

	file := todo.NewFile(todoPath)
	if err := tl.Load(file); err != nil {
		return err
	}

	// Add the todo (no parent, append to end)
	_, err := tl.Add(title, "", -1)
	if err != nil {
		return err
	}

	// Save to file
	return tl.Save(file)
}
