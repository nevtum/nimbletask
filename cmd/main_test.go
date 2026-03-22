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
