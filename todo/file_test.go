package todo

import (
	"path/filepath"
	"testing"
)

func TestFile_SaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "roundtrip.md")

	f := NewFile(tmpFile)
	content := "- [ ] <!-- id:1|created:2024-01-01T00:00:00Z --> Test\n  Description line"

	// Save
	err := f.Save(content)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load
	loaded, err := f.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded != content {
		t.Errorf("round-trip failed: expected %q, got %q", content, loaded)
	}
}
