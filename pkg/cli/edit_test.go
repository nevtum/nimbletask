package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nevtum/nimbletask/pkg/todo"
	"github.com/stretchr/testify/assert"
)

func TestEditCommand_UpdatesTitle(t *testing.T) {
	// 1. Setup: Create a temporary directory and a todo file
	tempDir, err := os.MkdirTemp("", "todo-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	todoFilePath := filepath.Join(tempDir, "test_todo.md")

	// Create an initial todo list with one item
	// We'll use the todo package directly to ensure a known state
	tl := todo.NewTodoList(todo.NewFile(todoFilePath))
	originalTitle := "Original Title"
	newTitle := "Updated Title"

	// Add a todo item
	item, err := tl.Add(originalTitle, "", -1)
	assert.NoError(t, err)

	// Save the initial state
	assert.NoError(t, tl.Save())

	// 2. Execute: Run the 'edit' command via the CLI
	rootCmd := NewRootCmd()
	// Tell the CLI to use our temporary file
	rootCmd.SetArgs([]string{"edit", item.ID, "--title", newTitle, "--file", todoFilePath})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err = rootCmd.Execute()

	// 3. Assert: Command should succeed
	assert.NoError(t, err)

	// 4. Verify: Reload the file and check the update
	updatedTl, _, err := loadTodoList(todoFilePath)
	assert.NoError(t, err)

	updatedItem, err := updatedTl.Get(item.ID)
	assert.NoError(t, err)

	assert.Equal(t, newTitle, updatedItem.Title)
}

func TestEditCommand_Failed(t *testing.T) {
	// 1. Setup: Create a temporary directory and a todo file
	tempDir, err := os.MkdirTemp("", "todo-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	todoFilePath := filepath.Join(tempDir, "test_todo.md")

	// 2. Execute: Run the 'edit' command via the CLI
	rootCmd := NewRootCmd()
	// Tell the CLI to use our temporary file
	rootCmd.SetArgs([]string{"edit", "id-does-not-exist", "--title", "new title", "--file", todoFilePath})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err = rootCmd.Execute()

	assert.Error(t, err)
}
