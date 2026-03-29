package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitConfigCommand tests that `todo init` creates the global configuration.
// This is the foundational command - all other operations depend on it.
func TestInitConfigCommand(t *testing.T) {
	// Use isolated temp directory (safe - no env vars modified)
	configDir := t.TempDir()

	_, err := runCmd(t, "--config", configDir, "init")

	// Should succeed without error
	require.NoError(t, err, "init should complete without error")

	// Verify config directory exists
	info, err := os.Stat(configDir)
	require.NoError(t, err, "config directory should be created")
	assert.True(t, info.IsDir(), "config path should be a directory")

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should be created")

	// Verify config file content
	configData, err := os.ReadFile(configPath)
	require.NoError(t, err, "should be able to read config file")
	assert.Contains(t, string(configData), `"filename"`, "config should contain filename field")
	assert.Contains(t, string(configData), `"todos.md"`, "config should contain default filename value")
}

// TestInitConfig_CreatesParentDirectories tests that init creates
// parent directories when they don't exist (fresh install scenario)
func TestInitConfig_CreatesParentDirectories(t *testing.T) {
	// Use a nested path that doesn't exist yet - simulating fresh install
	// where ~/.config/todos/ doesn't exist
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nested", ".config", "todos")

	// Ensure the path doesn't exist yet
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should not exist initially")

	_, err = runCmd(t, "--config", configDir, "init")

	require.NoError(t, err, "init should create parent directories and succeed")

	// Verify config directory was created
	info, err := os.Stat(configDir)
	require.NoError(t, err, "config directory should be created")
	assert.True(t, info.IsDir(), "config path should be a directory")

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should be created in new directory")

	// Verify config file content
	configData, err := os.ReadFile(configPath)
	require.NoError(t, err, "should be able to read config file")
	assert.Contains(t, string(configData), `"filename"`, "config should contain filename field")
	assert.Contains(t, string(configData), `"todos.md"`, "config should contain default filename value")
}

// TestInitConfig_Idempotent tests that init does NOT overwrite existing config
func TestInitConfig_Idempotent(t *testing.T) {
	configDir := t.TempDir()

	// Create the config directory
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "should create config directory")

	// Create a custom config file with non-default values
	customConfig, err := json.Marshal(Config{
		Filename: "custom.md",
	})
	require.NoError(t, err, "should marshal config")

	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, []byte(customConfig), 0644)
	require.NoError(t, err, "should create custom config file")

	// Run init - should NOT overwrite existing config
	out, err := runCmd(t, "--config", configDir, "init")

	// Command should succeed without error
	require.NoError(t, err, "init should complete without error")
	require.Contains(t, out.String(), "Config file already exists")

	// Verify config file content - should still have custom values
	configData, err := os.ReadFile(configPath)
	require.NoError(t, err, "should be able to read config file")

	// These assertions will FAIL with current implementation because it overwrites the file
	// The config should NOT be overwritten - custom values should be preserved
	configContent := string(configData)
	assert.Contains(t, configContent, `"custom.md"`, "init should NOT overwrite existing config")
}
