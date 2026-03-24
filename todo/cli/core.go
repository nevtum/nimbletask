package cli

import (
	"os"
	"path/filepath"

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

	// addCmd adds a new todo item to the list
	addCmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new todo item",
		RunE:  AddCmd(),
	}

	// initConfigCmd initializes the global configuration file
	initConfigCmd := &cobra.Command{
		Use:   "init-config",
		Short: "Initialize the global configuration file",
		RunE:  InitCmd(),
	}

	// Add subcommands to root
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(initConfigCmd)

	return rootCmd
}
