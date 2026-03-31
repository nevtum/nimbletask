package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteCommand_DeletesExistingTodo verifies that delete removes a todo from the list.
func TestDeleteCommand_DeletesExistingTodo(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Task to delete")
	require.NoError(t, err, "add command should succeed")

	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(string(content))
	require.NotEmpty(t, matches, "should find ID in file")
	todoID := matches[1]

	require.Contains(t, string(content), "Task to delete", "todo should exist before deletion")

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "delete", todoID)
	require.NoError(t, err, "delete should complete without error")

	content, err = os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file after delete")

	assert.NotContains(t, string(content), "Task to delete", "todo should be removed from file")
}

// TestDeleteCommand_ReturnsErrorForNonExistentID verifies that delete returns error for invalid ID.
func TestDeleteCommand_ReturnsErrorForNonExistentID(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Some task")
	require.NoError(t, err, "add command should succeed")

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "delete", "non-existent-id-123")

	assert.Error(t, err, "delete should return error for non-existent ID")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}

// TestDeleteCommand_MissingArgs verifies that delete command requires exactly one argument.
func TestDeleteCommand_MissingArgs(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "delete")

	assert.Error(t, err, "delete without arguments should error")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "error should mention argument count")
}

// TestDeleteCommand_NoConfigError verifies that delete returns error when config file doesn't exist.
func TestDeleteCommand_NoConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "delete", "some-id")

	assert.Error(t, err, "delete should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init must be called first", "error should mention init")
}

// TestDeleteCommand_RefusesToDeleteTodoWithChildren verifies delete returns error for parent todo.
func TestDeleteCommand_RefusesToDeleteTodoWithChildren(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Parent task")
	require.NoError(t, err, "add command should succeed for parent")

	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(string(content))
	require.NotEmpty(t, matches, "should find ID in file")
	parentID := matches[1]

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "--parent", parentID, "Child task")
	require.NoError(t, err, "add command should succeed for child")

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "delete", parentID)

	assert.Error(t, err, "delete should return error for todo with children")
	assert.Contains(t, err.Error(), "children", "error should mention children")
}

// TestDeleteCommand_WithForceFlagDeletesTodoWithChildren verifies --force flag allows deletion.
func TestDeleteCommand_WithForceFlagDeletesTodoWithChildren(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	setupTestConfig(t, tmpDir)

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Parent task")
	require.NoError(t, err, "add command should succeed for parent")

	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(string(content))
	require.NotEmpty(t, matches, "should find ID in file")
	parentID := matches[1]

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "--parent", parentID, "Child task")
	require.NoError(t, err, "add command should succeed for child")

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "delete", "--force", parentID)
	require.NoError(t, err, "delete with --force should succeed")

	content, err = os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file after delete")

	assert.NotContains(t, string(content), "Parent task", "parent todo should be removed from file")
	assert.Contains(t, string(content), "Child task", "child task should still exist")
}
