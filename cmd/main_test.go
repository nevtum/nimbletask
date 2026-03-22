package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitConfig_CreatesConfigFile is the FIRST test for CLI operations.
// It verifies that `todo init-config` creates the global configuration.
// This is the foundational command - all other operations depend on it.
func TestInitConfig_CreatesConfigFile(t *testing.T) {
	// Use isolated temp directory (safe - no env vars modified)
	tmpDir := t.TempDir()

	// Execute init-config with explicit config root (dependency injection)
	err := runInitConfig(tmpDir)

	// Should succeed without error
	require.NoError(t, err, "init-config should complete without error")

	// Verify config directory exists
	configDir := filepath.Join(tmpDir, ".todo")
	info, err := os.Stat(configDir)
	require.NoError(t, err, "config directory should be created")
	assert.True(t, info.IsDir(), "config path should be a directory")

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should be created")
}

// TestRunAdd is a table-driven test for the runAdd function.
// It covers both happy path and error scenarios.
func TestRunAdd(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		wantErr     bool
		errContains string
		wantContent string
	}{
		{
			name:        "creates todo with valid title",
			title:       "Buy groceries",
			wantErr:     false,
			wantContent: "Buy groceries",
		},
		{
			name:        "returns error for empty title",
			title:       "",
			wantErr:     true,
			errContains: "title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use isolated temp directory
			tmpDir := t.TempDir()

			// Initialize config first (prerequisite)
			err := runInitConfig(tmpDir)
			require.NoError(t, err, "init-config should complete without error")

			// Attempt to add todo
			todoPath := filepath.Join(tmpDir, "todos.md")
			err = runAdd(tmpDir, tt.title, todoPath)

			// Verify error expectations
			if tt.wantErr {
				assert.Error(t, err, "add should return error")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "error message should contain expected text")
				}
				return
			}

			// Verify success expectations
			require.NoError(t, err, "add should complete without error")

			// Verify the todo list file exists
			_, err = os.Stat(todoPath)
			assert.NoError(t, err, "todo list file should be created")

			// Verify the file contains the expected content
			content, err := os.ReadFile(todoPath)
			require.NoError(t, err, "should be able to read todo file")
			assert.Contains(t, string(content), tt.wantContent, "todo file should contain expected content")
		})
	}
}

// TestCLI_AddCommand tests the CLI-level execution with command arguments.
// It verifies that running "todo add Title" creates a todo in the list.
func TestCLI_AddCommand(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()

	// Initialize config first (prerequisite)
	err := runInitConfig(tmpDir)
	require.NoError(t, err, "init-config should complete without error")

	// Execute CLI command: todo add "Buy groceries"
	args := []string{"add", "Buy groceries"}
	err = runCLI(args, tmpDir)

	// Should succeed without error
	require.NoError(t, err, "CLI command should complete without error")

	// Verify the todo list file exists
	todoPath := filepath.Join(tmpDir, "todos.md")
	_, err = os.Stat(todoPath)
	assert.NoError(t, err, "todo list file should be created")

	// Verify the file contains the todo
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")
	assert.Contains(t, string(content), "Buy groceries", "todo file should contain the todo title")
}
