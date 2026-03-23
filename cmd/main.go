package main

import (
	"fmt"
	"os"
	"path/filepath"

	"todo_cli/todo"

	"github.com/spf13/cobra"
)

// Global variables to store flag values
var (
	configRoot string
	todoPath   string
)

// rootCmd is the main entry point for the CLI
var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "A CLI tool for managing hierarchical todo lists",
	Long:  `Manage your todo lists through a powerful CLI interface.`,
}

// addCmd adds a new todo item to the list
var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new todo item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		// Determine todoPath if not set via flag
		path := todoPath
		if path == "" {
			path = filepath.Join(configRoot, "todos.md")
		}
		return runAdd(configRoot, title, path)
	},
}

// initConfigCmd initializes the global configuration file
var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Initialize the global configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInitConfig(configRoot)
	},
}

// init registers subcommands and sets up flags
func init() {
	// Determine default config root (user's config directory)
	defaultConfigRoot := filepath.Join(os.Getenv("HOME"), ".config", "todo")

	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&configRoot, "config", defaultConfigRoot, "Configuration directory root")
	rootCmd.PersistentFlags().StringVar(&todoPath, "file", "", "Path to todo list file (default: <config>/todos.md)")

	// Add subcommands to root
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(initConfigCmd)
}

// main executes the root command
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// runCLI parses and executes CLI commands
// This function is maintained for backward compatibility with tests
func runCLI(args []string, cfgRoot string) error {
	// Set up the command with the provided config root
	configRoot = cfgRoot
	todoPath = filepath.Join(configRoot, "todos.md")

	if len(args) < 2 {
		return fmt.Errorf("usage: add [title]")
	}

	title := args[1]
	return runAdd(configRoot, title, todoPath)
}

// runInitConfig creates the configuration directory and file
func runInitConfig(configRoot string) error {
	configDir := filepath.Join(configRoot, ".todo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	return os.WriteFile(configPath, []byte("{}"), 0644)
}

// runAdd creates a new todo and saves it to the todo list file
func runAdd(configRoot string, title string, todoPath string) error {
	// Create a new todo list
	tl := todo.NewTodoList()

	// Add the todo (no parent, append to end)
	_, err := tl.Add(title, "", -1)
	if err != nil {
		return err
	}

	// Save to file
	file := todo.NewFile(todoPath)
	return tl.Save(file)
}
