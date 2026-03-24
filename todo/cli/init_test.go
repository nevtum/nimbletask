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
