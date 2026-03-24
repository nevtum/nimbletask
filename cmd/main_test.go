package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitConfigCommand tests that `todo init-config` creates the global configuration.
// This is the foundational command - all other operations depend on it.
func TestInitConfigCommand(t *testing.T) {
	// Use isolated temp directory (safe - no env vars modified)
	configDir := t.TempDir()

	// Get a fresh command instance
	cmd := NewRootCmd()

	// Set arguments: init-config with custom config directory
	cmd.SetArgs([]string{"--config", configDir, "init-config"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()

	// Should succeed without error
	require.NoError(t, err, "init-config should complete without error")

	// Verify config directory exists
	info, err := os.Stat(configDir)
	require.NoError(t, err, "config directory should be created")
	assert.True(t, info.IsDir(), "config path should be a directory")

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should be created")
}

// TestAddCommand is a table-driven test for the `add` cobra command.
// It covers both happy path and error scenarios.
func TestAddCommand(t *testing.T) {
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
			todoFile := filepath.Join(tmpDir, "todo.md")

			// Get a fresh command instance for add with custom file path
			addCmd := NewRootCmd()
			addCmd.SetArgs([]string{"--file", todoFile, "add", tt.title})

			// Capture output
			var out bytes.Buffer
			addCmd.SetOut(&out)
			addCmd.SetErr(&out)

			// Execute add command
			err := addCmd.Execute()

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
			_, err = os.Stat(todoFile)
			assert.NoError(t, err, "todo list file should be created")

			// Verify the file contains the expected content
			content, err := os.ReadFile(todoFile)
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

	// Execute CLI command: todo add "Buy groceries"
	// Use --file flag to specify test-specific location
	todoPath := filepath.Join(tmpDir, "todo.md")
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--file", todoPath, "add", "Buy groceries"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()

	// Should succeed without error
	require.NoError(t, err, "CLI command should complete without error")

	// Verify the todo list file exists
	_, err = os.Stat(todoPath)
	assert.NoError(t, err, "todo list file should be created")

	// Verify the file contains the todo
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")
	assert.Contains(t, string(content), "Buy groceries", "todo file should contain the todo title")
}

// TestAddCommand_ExactArgs tests that add command requires exactly one argument
func TestAddCommand_ExactArgs(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	// Get a fresh command instance
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--file", todoPath, "add"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute without arguments
	err := cmd.Execute()

	// Should error due to missing argument
	assert.Error(t, err, "add without arguments should error")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "error should mention argument count")
}

// TestRootCmd_NoSubcommand tests that running `todo` without subcommand shows help
func TestRootCmd_NoSubcommand(t *testing.T) {
	// Get a fresh command instance
	cmd := NewRootCmd()

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute without subcommand
	err := cmd.Execute()

	// Should not error (shows help)
	assert.NoError(t, err, "root command without args should show help")

	// Verify help is shown
	output := buf.String()
	assert.Contains(t, output, "Usage:", "help should contain usage information")
	assert.Contains(t, output, "Available Commands:", "help should list available commands")
}

// TestAddCommand_WithCustomFileFlag tests the add command with --file flag
func TestAddCommand_WithCustomFileFlag(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	customFile := filepath.Join(tmpDir, "custom-todos.md")

	// Execute add with custom file path
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--file", customFile, "add", "Custom file todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()
	require.NoError(t, err, "add with custom file should succeed")

	// Verify the custom file exists
	_, err = os.Stat(customFile)
	assert.NoError(t, err, "custom todo file should be created")

	// Verify content
	content, err := os.ReadFile(customFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Custom file todo")
}

// TestAddCommand_DefaultPWD tests that add command uses todo.md in current directory when no --file flag is specified
func TestAddCommand_DefaultPWD(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()

	// Change to temp directory and restore afterwards
	origDir, err := os.Getwd()
	require.NoError(t, err, "should get current directory")
	err = os.Chdir(tmpDir)
	require.NoError(t, err, "should change to temp directory")
	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err, "should restore original directory")
	}()

	// Execute add without --file flag (should default to todo.md in PWD)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"add", "Default PWD todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err = cmd.Execute()
	require.NoError(t, err, "add without --file flag should succeed")

	// Verify the todo.md file exists in the current directory (tmpDir)
	expectedFile := filepath.Join(tmpDir, "todo.md")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err, "todo.md should be created in current directory")

	// Verify content
	content, err := os.ReadFile(expectedFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Default PWD todo")
}

// TestInitConfig_Idempotent tests that init-config can be run multiple times
func TestInitConfig_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// First execution
	cmd1 := NewRootCmd()
	cmd1.SetArgs([]string{"--config", tmpDir, "init-config"})
	err := cmd1.Execute()
	require.NoError(t, err, "first init-config should succeed")

	// Second execution - should not fail
	cmd2 := NewRootCmd()
	cmd2.SetArgs([]string{"--config", tmpDir, "init-config"})
	err = cmd2.Execute()
	require.NoError(t, err, "second init-config should also succeed (idempotent)")

	// Verify config still exists
	configPath := filepath.Join(tmpDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should still exist")
}
