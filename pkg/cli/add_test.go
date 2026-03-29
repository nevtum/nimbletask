package cli

import (
	"bytes"
	"fmt"
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

			_, err := runCmd(t, "--config", tmpDir, "--file", todoFile, "add", tt.title)

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
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Buy groceries")

	// Should succeed without error
	require.NoError(t, err, "CLI command should complete without error")
	assert.Contains(t, out.String(), "Todo created", "output should contain a success message")

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

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add")

	// Should error due to missing argument
	assert.Error(t, err, "add without arguments should error")
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0", "error should mention argument count")
}

// TestRootCmd_NoSubcommand tests that running `todo` without subcommand shows help
func TestRootCmd_NoSubcommand(t *testing.T) {
	buf, err := runCmd(t)

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
	_, err := runCmd(t, "--config", tmpDir, "--file", customFile, "add", "Custom file todo")
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
	_, err = runCmd(t, "--config", tmpDir, "add", "Default PWD todo")
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

	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Test todo")

	// Should error due to missing config
	assert.Error(t, err, "add should return error when config doesn't exist")
	assert.Contains(t, err.Error(), "init must be called first", "error should mention init")
}

// TestAddCommand_NoConfig_NoUsage tests that when config is missing, only error is shown (no usage docs)
func TestAddCommand_NoConfig_NoUsage(t *testing.T) {
	// Use isolated temp directory (but don't create config)
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Execute command with runCmd helper - no config setup
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Test todo")

	// Should error
	require.Error(t, err, "add should return error when config doesn't exist")

	// Verify error message is shown
	output := out.String()
	assert.Contains(t, output, "init must be called first", "error message should be shown")

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

	// Execute without arguments using runCmd helper
	out, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add")

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

	// Execute command using runCmd helper
	_, err = runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Test todo")

	// Should error due to malformed config
	require.Error(t, err, "add should return error when config file has malformed JSON")
	assert.Contains(t, err.Error(), "failed to parse config file", "error should mention config file parsing")
}

// TestAddCommand_UsesDefaultPriorityFromConfig verifies that the add command
// reads default_priority from config and applies it to new todos.
// This is foundational - config values should drive default behavior.
func TestAddCommand_UsesDefaultPriorityFromConfig(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file with default_priority: 3
	setupTestConfigWithPriority(t, tmpDir, 3)

	// Execute add without --priority flag (should use default from config)
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Priority test todo")
	require.NoError(t, err, "add should complete without error")

	// Verify the todo list file exists
	_, err = os.Stat(todoPath)
	assert.NoError(t, err, "todo list file should be created")

	// Verify the file contains the expected priority in metadata
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	// The metadata should contain priority:3 from config default
	contentStr := string(content)
	assert.Contains(t, contentStr, "Priority test todo", "todo file should contain the todo title")
	assert.Contains(t, contentStr, "priority:3", "todo should have default priority from config")
}

// TestAddCommand_WithPriorityFlag verifies that the --priority CLI flag
// overrides the default_priority from config when creating a new todo.
// This ensures CLI flags take precedence over configuration defaults.
func TestAddCommand_WithPriorityFlag(t *testing.T) {
	// Use isolated temp directory
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todos.md")

	// Setup config file with default_priority: 3
	setupTestConfigWithPriority(t, tmpDir, 3)

	// Execute add with --priority flag (should override config default)
	_, err := runCmd(t, "--config", tmpDir, "--file", todoPath, "add", "Priority flag todo", "--priority", "5")
	require.NoError(t, err, "add with --priority flag should succeed")

	// Verify the todo list file exists
	_, err = os.Stat(todoPath)
	assert.NoError(t, err, "todo list file should be created")

	// Verify the file contains the CLI-specified priority (not the config default)
	content, err := os.ReadFile(todoPath)
	require.NoError(t, err, "should be able to read todo file")

	contentStr := string(content)
	assert.Contains(t, contentStr, "Priority flag todo", "todo file should contain the todo title")
	assert.Contains(t, contentStr, "priority:5", "todo should have priority from CLI flag (5), not config default (3)")
	assert.NotContains(t, contentStr, "priority:3", "todo should NOT have config default priority when flag is provided")
}

// setupTestConfigWithPriority creates a config file with specified default_priority
func setupTestConfigWithPriority(t *testing.T, configDir string, priority int) {
	t.Helper()
	configPath := filepath.Join(configDir, "config.json")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "should create config directory")
	configContent := fmt.Sprintf(`{"filename": "todos.md", "default_priority": %d}`, priority)
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "should create config file with default_priority")
}

// runCmd runs the CLI command with the given arguments
// and returns the output buffer and error
func runCmd(t *testing.T, args ...string) (bytes.Buffer, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out, err
}
