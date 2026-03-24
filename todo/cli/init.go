package cli

import (
	"os"
	"path/filepath"
)

// runInitConfig creates the configuration directory and file
func runInitConfig(configRoot string) error {
	configPath := filepath.Join(configRoot, "config.json")
	return os.WriteFile(configPath, []byte("{}"), 0644)
}
