package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig creates a config file at the specified config directory
func setupTestConfig(t *testing.T, configDir string) {
	t.Helper()
	configPath := filepath.Join(configDir, "config.json")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "should create config directory")
	err = os.WriteFile(configPath, []byte(`{"filename": "todos.md"}`), 0644)
	require.NoError(t, err, "should create config file")
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
			todoFile := filepath.Join(tmpDir, "todos.md")

			// Setup config file first
			setupTestConfig(t, tmpDir)

			// Get a fresh command instance for add with custom file path
			addCmd := NewRootCmd()
			addCmd.SetArgs([]string{"--config", tmpDir, "--file", todoFile, "add", tt.title})

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

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Execute CLI command: todo add "Buy groceries"
	// Use --file flag to specify test-specific location
	todoPath := filepath.Join(tmpDir, "todos.md")
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", "Buy groceries"})

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
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Get a fresh command instance
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add"})

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

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Execute add with custom file path
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", customFile, "add", "Custom file todo"})

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

	// Setup config file first
	setupTestConfig(t, tmpDir)

	// Change to temp directory and restore afterwards
	origDir, err := os.Getwd()
	require.NoError(t, err, "should get current directory")
	err = os.Chdir(tmpDir)
	require.NoError(t, err, "should change to temp directory")
	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err, "should restore original directory")
	}()

	// Execute add without --file flag (should default to todos.md from config in PWD)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "add", "Default PWD todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err = cmd.Execute()
	require.NoError(t, err, "add without --file flag should succeed")

	// Verify the todos.md file exists in the current directory (tmpDir)
	expectedFile := filepath.Join(tmpDir, "todos.md")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err, "todos.md should be created in current directory")

	// Verify content
	content, err := os.ReadFile(expectedFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Default PWD todo")
}

// TestAddCommand_NoConfigError tests that add returns error when config file doesn't exist
func TestAddCommand_NoConfigError(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Get a fresh command instance - no config setup
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", "Test todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()

	// Should error due to missing config
	assert.Error(t, err, "add should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init-config must be called first", "error should mention init-config")
}

// TestAddCommand_NoConfig_NoUsage tests that when config is missing, only error is shown (no usage docs)
func TestAddCommand_NoConfig_NoUsage(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Get a fresh command instance - no config setup
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", "Test todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err := cmd.Execute()

	// Should error
	require.Error(t, err, "add should return error when config doesn't exist")

	// Verify error message is shown
	output := out.String()
	assert.Contains(t, output, "init-config must be called first", "error message should be shown")

	// Verify usage is NOT shown (SilenceUsage should prevent this)
	assert.NotContains(t, output, "Usage:", "usage should NOT be shown when config is missing")
	assert.NotContains(t, output, "Flags:", "flags should NOT be shown when config is missing")
}

// TestAddCommand_MissingArgs_ShowsUsage tests that when args are missing, usage docs are shown
func TestAddCommand_MissingArgs_ShowsUsage(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file first (so config error doesn't mask the args error)
	setupTestConfig(t, tmpDir)

	// Get a fresh command instance - no title argument
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute without arguments
	err := cmd.Execute()

	// Should error due to missing argument
	require.Error(t, err, "add without arguments should error")

	// Verify usage IS shown for missing arguments (SilenceUsage only applies to RunE errors, not arg validation)
	output := out.String()
	assert.Contains(t, output, "Usage:", "usage should be shown when arguments are missing")
}

// TestAddCommand_MalformedConfig tests that add returns error when config file contains invalid JSON
func TestAddCommand_MalformedConfig(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Create config directory
	configPath := filepath.Join(tmpDir, "config.json")
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err, "should create config directory")

	// Write malformed JSON to config file (simulating user editing error)
	err = os.WriteFile(configPath, []byte(`{invalid json}`), 0644)
	require.NoError(t, err, "should create malformed config file")

	// Get a fresh command instance
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", tmpDir, "--file", todoPath, "add", "Test todo"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Execute command
	err = cmd.Execute()

	// Should error due to malformed config
	require.Error(t, err, "add should return error when config file has malformed JSON")
	assert.Contains(t, err.Error(), "failed to parse config file", "error should mention config file parsing")
}
