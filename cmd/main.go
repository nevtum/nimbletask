package main

import (
	"os"
	"path/filepath"
)

func main() {
	// TODO: implement CLI entry point
}

// runInitConfig creates the configuration directory and file
func runInitConfig(configRoot string) error {
	configDir := filepath.Join(configRoot, ".todo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	return os.WriteFile(configPath, []byte("{}"), 0644)
}
