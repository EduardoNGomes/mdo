package theme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Option struct {
	ID    string
	Name  string
	Light bool
}

var Options = []Option{
	{"default", "Default", false},
	{"default-light", "Default light", true},
	{"catppuccin-latte", "Catppuccin Latte", true},
	{"catppuccin-frappe", "Catppuccin Frappé", false},
	{"catppuccin-macchiato", "Catppuccin Macchiato", false},
	{"catppuccin-mocha", "Catppuccin Mocha", false},
	{"nord", "Nord", false},
	{"tokyo-night", "Tokyo Night", false},
	{"tokyo-night-storm", "Tokyo Night Storm", false},
	{"tokyo-night-moon", "Tokyo Night Moon", false},
	{"tokyo-night-day", "Tokyo Night Day", true},
	{"dracula-classic", "Dracula Classic", false},
	{"alucard-classic", "Alucard Classic", true},
}

func Lookup(id string) (Option, bool) {
	for _, option := range Options {
		if option.ID == id {
			return option, true
		}
	}
	return Option{}, false
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mdo", "config.json"), nil
}

func Load(path string) (string, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read theme config %q: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 4097))
	if err != nil {
		return "", fmt.Errorf("read theme config %q: %w", path, err)
	}
	if len(data) > 4096 {
		return "", fmt.Errorf("theme config %q is too large", path)
	}
	var config struct {
		Theme string `json:"theme"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return "", fmt.Errorf("parse theme config %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return "", fmt.Errorf("parse theme config %q: trailing data", path)
	}
	if _, ok := Lookup(config.Theme); !ok {
		return "", fmt.Errorf("theme config %q has unknown theme %q", path, config.Theme)
	}
	return config.Theme, nil
}

func Save(path, id string) error {
	if _, ok := Lookup(id); !ok {
		return fmt.Errorf("unknown theme %q", id)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create theme config directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure theme config directory: %w", err)
	}
	temp, err := os.CreateTemp(dir, ".config-*")
	if err != nil {
		return fmt.Errorf("create theme config: %w", err)
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return err
	}
	data, _ := json.Marshal(struct {
		Theme string `json:"theme"`
	}{id})
	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return fmt.Errorf("write theme config: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close theme config: %w", err)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("save theme config: %w", err)
	}
	return nil
}
