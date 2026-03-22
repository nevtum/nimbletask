package main

import (
	"os"
	"path/filepath"

	"todo_cli/todo"
)

func main() {
	// TODO: implement CLI entry point
}

// runCLI parses and executes CLI commands
func runCLI(args []string, configRoot string) error {
	title := args[1]
	todoPath := filepath.Join(configRoot, "todos.md")
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
