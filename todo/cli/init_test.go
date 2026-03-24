package cli

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

// TestInitConfig_CreatesParentDirectories tests that init-config creates
// parent directories when they don't exist (fresh install scenario)
func TestInitConfig_CreatesParentDirectories(t *testing.T) {
	// Use a nested path that doesn't exist yet - simulating fresh install
	// where ~/.config/todos/ doesn't exist
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nested", ".config", "todos")

	// Ensure the path doesn't exist yet
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should not exist initially")

	// Execute init-config
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--config", configDir, "init-config"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err = cmd.Execute()
	require.NoError(t, err, "init-config should create parent directories and succeed")

	// Verify config directory was created
	info, err := os.Stat(configDir)
	require.NoError(t, err, "config directory should be created")
	assert.True(t, info.IsDir(), "config path should be a directory")

	// Verify config file exists
	configPath := filepath.Join(configDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config.json should be created in new directory")
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
