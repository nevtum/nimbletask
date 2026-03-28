package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
// TestListCommand_DisplaysTodos verifies that the list command displays todos
// in the numbered path format as specified. This is the foundational test for
// the list command - validating that users can view their hierarchical todos.
//
// According to the spec, the default display format should be:
//  1. [ ] V1StGXR8_Z5jd Project Proposal
//     1.1. [x] wH9mK2pL4nQ7 Research
//
// This test establishes the core read operation that all other list features
// build upon.
func TestListCommand_DisplaysTodos(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Setup: Create multiple todos to test hierarchy display
	// First, create root-level todos
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Project Proposal")
	require.NoError(t, err, "setup: first add command should succeed")

	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Second Task")
	require.NoError(t, err, "setup: second add command should succeed")

	// Execute: Run list command
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "list")

	// Verify: Command should succeed
	require.NoError(t, err, "list command should complete without error")

	// Verify: Output should contain expected elements
	output := out.String()

	// Should show numbered paths (1., 2., etc.)
	assert.Regexp(t, regexp.MustCompile(`(?m)^\s*1\.`), output, "output should contain numbered path '1.'")
	assert.Regexp(t, regexp.MustCompile(`(?m)^\s*2\.`), output, "output should contain numbered path '2.'")

	// Should show checkboxes for incomplete todos
	assert.Contains(t, output, "[ ]", "output should contain incomplete checkbox [ ]")

	// Should show todo titles
	assert.Contains(t, output, "Project Proposal", "output should contain first todo title")
	assert.Contains(t, output, "Second Task", "output should contain second todo title")
}
*/

// TestListCommand_NoConfigError tests that list returns error when config file doesn't exist
func TestListCommand_NoConfigError(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Run command - no config setup
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "list")

	// Should error due to missing config
	assert.Error(t, err, "list should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init-config must be called first", "error should mention init-config")
}

/*
// TestListCommand_EmptyList tests that list handles empty todo list gracefully
func TestListCommand_EmptyList(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Execute: Run list command on empty list
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "list")

	// Verify: Command should succeed even with empty list
	require.NoError(t, err, "list command should succeed with empty list")

	// Output should be empty or indicate no todos
	output := out.String()
	// Empty output is acceptable, or a message like "No todos found"
	_ = output
}
*/
