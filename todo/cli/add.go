package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"todo_cli/todo"
)

// runAdd creates a new todo and saves it to the todo list file
func runAdd(configRoot string, todoPath string, args ...string) error {
	// Check if config file exists - init-config must be called first
	configPath := filepath.Join(configRoot, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found at %s: init-config must be called first", configPath)
	}

	// Determine todoPath if not set via flag - default to todo.md in PWD
	if todoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		todoPath = filepath.Join(cwd, "todo.md")
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
	_, err := tl.Add(title, "", -1)
	if err != nil {
		return err
	}

	// Save to file
	return tl.Save(file)
}
