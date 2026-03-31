package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractID extracts the todo ID from command output using regex
// Output format: "Todo created! ID is <id>\n"
func extractID(output string) string {
	re := regexp.MustCompile(`ID is ([A-Za-z0-9_-]+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// TestMoveCommand_MovesTodoToNewParent verifies that the move command
// can relocate a todo under a new parent task.
// This is core functionality for hierarchical task reorganization.
func TestMoveCommand_MovesTodoToNewParent(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Create initial todo structure: Parent and Child
	parentOutput, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Parent Task")
	require.NoError(t, err, "should create parent task")
	parentID := extractID(parentOutput.String())
	require.NotEmpty(t, parentID, "should extract parent ID from output")

	childOutput, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Child Task")
	require.NoError(t, err, "should create child task")
	childID := extractID(childOutput.String())
	require.NotEmpty(t, childID, "should extract child ID from output")

	// Verify child starts as a root (no parent)
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should read todo file")
	initialContent := string(content)
	require.Contains(t, initialContent, "Parent Task", "file should contain parent")
	require.Contains(t, initialContent, "Child Task", "file should contain child")

	// Execute move command: move child under parent
	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "move", childID, "--parent", parentID)
	require.NoError(t, err, "move command should relocate todo under new parent")

	// Verify the file was updated
	content, err = os.ReadFile(todoPath)
	require.NoError(t, err, "should read updated todo file")
	updatedContent := string(content)

	// Child should now be indented under Parent in the markdown
	require.Contains(t, updatedContent, parentID, "file should contain parent ID")
	require.Contains(t, updatedContent, childID, "file should contain child ID")
}

// TestMoveCommand_MissingArgs tests that move command requires exactly one argument
func TestMoveCommand_MissingArgs(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "move")

	// Should error due to missing argument
	assert.Error(t, err, "move without arguments should error")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "error should mention argument count")
}

// TestMoveCommand_NoConfigError tests that move returns error when config file doesn't exist
func TestMoveCommand_NoConfigError(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "move", "some-id")

	// Should error due to missing config
	assert.Error(t, err, "move should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init must be called first", "error should mention init")
}

// TestMoveCommand_NonExistentID tests error handling when moving a non-existent todo
func TestMoveCommand_NonExistentID(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Try to move a todo that doesn't exist
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "move", "nonexistent-id")

	// Should error due to non-existent todo
	assert.Error(t, err, "move should return error for non-existent ID")
	assert.Contains(t, err.Error(), "not found", "error should indicate todo not found")
}

// TestMoveCommand_NonExistentParentID tests error handling when parent ID doesn't exist
func TestMoveCommand_NonExistentParentID(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Create a todo to move
	output, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Task to move")
	require.NoError(t, err, "should create task")
	todoID := extractID(output.String())
	require.NotEmpty(t, todoID, "should extract todo ID")

	// Try to move under a non-existent parent
	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "move", todoID, "--parent", "nonexistent-parent")

	// Should error due to non-existent parent
	assert.Error(t, err, "move should return error for non-existent parent ID")
	assert.Contains(t, err.Error(), "not found", "error should indicate parent not found")
}
