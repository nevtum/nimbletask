package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	output := out.String()
	// Output should indicate no todos
	assert.Equal(t, "No todos found\n", output)
}

// TestListCommand_HierarchicalDisplay verifies that the list command correctly displays
// hierarchical todos with numbered paths (1.1, 1.2, etc.) as specified in the
// display format requirements.
//
// This test is foundational for validating parent-child relationships work end-to-end.
func TestListCommand_HierarchicalDisplay(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Setup: Create parent todo
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Project Proposal")
	require.NoError(t, err, "setup: add parent todo should succeed")

	// Extract parent ID from file
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "setup: should be able to read todo file")
	re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(string(content))
	require.NotEmpty(t, matches, "setup: should find ID in file")
	parentID := matches[1]

	// Setup: Create child todo with --parent flag
	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Research", "--parent", parentID)
	require.NoError(t, err, "setup: add child todo should succeed")

	// Execute: Run list command
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "list")

	// Verify: Command should succeed
	require.NoError(t, err, "list command should complete without error")

	// Verify: Output should contain hierarchical numbering
	output := out.String()

	// Should show parent as "1."
	assert.Regexp(t, regexp.MustCompile(`(?m)^\s*1\.\s+\[\s*]`), output, "output should contain parent numbered '1.'")

	// Should show child as "1.1" (hierarchical numbering)
	assert.Regexp(t, regexp.MustCompile(`(?m)^\s*1\.1\s+\[\s*]`), output, "output should contain child numbered '1.1'")

	// Should show both titles
	assert.Contains(t, output, "Project Proposal", "output should contain parent title")
	assert.Contains(t, output, "Research", "output should contain child title")
}
