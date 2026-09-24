package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "mdo", "config.json")
	if got, err := Load(path); err != nil || got != "" {
		t.Fatalf("Load(missing) = %q, %v", got, err)
	}
	for _, id := range []string{"catppuccin-latte", "tokyo-night-moon", "alucard-classic"} {
		if err := Save(path, id); err != nil {
			t.Fatal(err)
		}
		if got, err := Load(path); err != nil || got != id {
			t.Fatalf("Load() = %q, %v; want %q", got, err, id)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config permissions = %o, want 600", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Errorf("config directory permissions = %o, want 700", dirInfo.Mode().Perm())
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	for _, content := range []string{`{"theme":"unknown"}`, `{"theme":`, `{"theme":"default"} {}`} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Errorf("Load(%q) accepted invalid config", content)
		}
	}
	if err := Save(path, "unknown"); err == nil || !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("Save(unknown) error = %v", err)
	}
}

func TestOptionsHaveUniqueIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, option := range Options {
		if seen[option.ID] || option.ID == "" || option.Name == "" {
			t.Fatalf("duplicate or empty theme option: %+v", option)
		}
		seen[option.ID] = true
	}
	if len(Options) != 13 {
		t.Fatalf("got %d theme options, want 13", len(Options))
	}
}
