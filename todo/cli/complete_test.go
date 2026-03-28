package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteCommand is a table-driven test for the `complete` cobra command.
// It verifies that running `todo complete <id>` marks a todo as completed
// and persists the change to the markdown file.
func TestCompleteCommand(t *testing.T) {
	tests := []struct {
		name          string
		setupTodos    []string // Titles to pre-populate
		targetIndex   int      // Which todo to complete (by index), -1 for invalid ID
		useInvalidID  bool     // Whether to use a non-existent ID
		wantErr       bool
		errContains   string
		wantCompleted bool // Whether the target should be marked complete
	}{
		{
			name:          "marks existing todo as complete",
			setupTodos:    []string{"Task 1", "Task 2"},
			targetIndex:   0,
			wantErr:       false,
			wantCompleted: true,
		},
		{
			name:         "returns error for non-existent ID",
			setupTodos:   []string{"Task 1"},
			targetIndex:  -1,
			useInvalidID: true,
			wantErr:      true,
			errContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use isolated temp directory
			tmpDir := t.TempDir()
			todoPath := filepath.Join(tmpDir, "todos.md")

			// Setup config file first
			setupTestConfig(t, tmpDir)

			// Setup: Create todo file with pre-existing todos
			var targetID string
			if len(tt.setupTodos) > 0 {
				// Create todos using add command
				for i, title := range tt.setupTodos {
					cmd := NewRootCmd()
					cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", title})

					var out bytes.Buffer
					cmd.SetOut(&out)
					cmd.SetErr(&out)

					err := cmd.Execute()
					require.NoError(t, err, "setup: add command should succeed for todo %d", i)
				}

				// Extract first ID from file for target todo
				if tt.targetIndex >= 0 {
					content, err := os.ReadFile(todoPath)
					require.NoError(t, err, "should be able to read todo file")

					// Extract ID using regex: id:([a-zA-Z0-9_-]+)
					re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
					matches := re.FindAllStringSubmatch(string(content), -1)
					require.NotEmpty(t, matches, "should find at least one ID in file")

					if tt.targetIndex < len(matches) {
						targetID = matches[tt.targetIndex][1]
					}
				}
			}

			// Determine the ID to use for complete command
			completeID := targetID
			if tt.useInvalidID {
				completeID = "invalid-id-123"
			}

			// Execute: Run complete command
			completeCmd := NewRootCmd()
			completeCmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "complete", completeID})

			// Capture output
			var out bytes.Buffer
			completeCmd.SetOut(&out)
			completeCmd.SetErr(&out)

			// Execute command
			err := completeCmd.Execute()

			// Verify error expectations
			if tt.wantErr {
				assert.Error(t, err, "complete should return error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "error message should contain expected text")
				}
				return
			}

			// Verify success expectations
			require.NoError(t, err, "complete should complete without error")

			// Verify the file was updated with completed status
			content, err := os.ReadFile(todoPath)
			require.NoError(t, err, "should be able to read todo file")

			contentStr := string(content)

			// For the target todo, verify it has [x] (completed) not just [ ]
			if tt.wantCompleted && len(tt.setupTodos) > 0 && tt.targetIndex >= 0 {
				// Count occurrences of [x] - should be at least 1
				assert.Contains(t, contentStr, "[x]", "todo file should contain completed checkbox")

				// Also verify the specific todo title is in the file with completed status
				// The completed todo should have [x] before the title
				targetTitle := tt.setupTodos[tt.targetIndex]
				// Look for a line with [x] that contains the target title
				lines := bytes.Split(content, []byte("\n"))
				foundCompleted := false
				for _, line := range lines {
					if bytes.Contains(line, []byte("[x]")) && bytes.Contains(line, []byte(targetTitle)) {
						foundCompleted = true
						break
					}
				}
				assert.True(t, foundCompleted, "todo '%s' should be marked as complete", targetTitle)
			}
		})
	}
}

// TestCompleteCommand_MissingArgs tests that complete command requires exactly one argument (the ID)
func TestCompleteCommand_MissingArgs(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Get a fresh command instance - no ID argument
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "complete"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute without arguments
	err := cmd.Execute()

	// Should error due to missing argument
	assert.Error(t, err, "complete without arguments should error")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "error should mention argument count")
}

// TestCompleteCommand_NoConfigError tests that complete returns error when config file doesn't exist
func TestCompleteCommand_NoConfigError(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Get a fresh command instance - no config setup
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "complete", "some-id"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()

	// Should error due to missing config
	assert.Error(t, err, "complete should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init-config must be called first", "error should mention init-config")
}

/*
// TestCompleteCommand_AlreadyCompleted tests that completing an already completed todo succeeds (idempotent)
func TestCompleteCommand_AlreadyCompleted(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Setup: Create a todo
	addCmd := NewRootCmd()
	addCmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", "Already done task"})

	var out bytes.Buffer
	addCmd.SetOut(&out)
	addCmd.SetErr(&out)

	err := addCmd.Execute()
	require.NoError(t, err, "add command should succeed")

	// Extract the ID from the file
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	re := regexp.MustCompile(`id:([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(string(content))
	require.NotEmpty(t, matches, "should find ID in file")
	todoID := matches[1]

	// First complete
	completeCmd1 := NewRootCmd()
	completeCmd1.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "complete", todoID})
	err = completeCmd1.Execute()
	require.NoError(t, err, "first complete should succeed")

	// Second complete (already completed)
	completeCmd2 := NewRootCmd()
	completeCmd2.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "complete", todoID})
	err = completeCmd2.Execute()

	// Should succeed (idempotent)
	assert.NoError(t, err, "completing already-completed todo should succeed")
}
*/
