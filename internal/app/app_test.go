package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMarkdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "document.md")
	const content = "# Documento\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	absPath, data, err := readMarkdown(path)
	if err != nil {
		t.Fatalf("readMarkdown() error = %v", err)
	}
	if !filepath.IsAbs(absPath) {
		t.Errorf("readMarkdown() path = %q, want absolute path", absPath)
	}
	if string(data) != content {
		t.Errorf("readMarkdown() data = %q, want %q", data, content)
	}
}

func TestReadMarkdownRejectsInvalidInputs(t *testing.T) {
	dir := t.TempDir()
	txtPath := filepath.Join(dir, "document.txt")
	if err := os.WriteFile(txtPath, []byte("text"), 0o600); err != nil {
		t.Fatal(err)
	}
	markdownDir := filepath.Join(dir, "folder.md")
	if err := os.Mkdir(markdownDir, 0o700); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "wrong extension", path: txtPath, wantErr: "extensão .md"},
		{name: "uppercase extension", path: filepath.Join(dir, "DOCUMENT.MD"), wantErr: "extensão .md"},
		{name: "directory", path: markdownDir, wantErr: "não é um arquivo regular"},
		{name: "missing file", path: filepath.Join(dir, "missing.md"), wantErr: "no such file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := readMarkdown(tt.path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("readMarkdown() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestReadMarkdownRejectsLargeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.md")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(maxMarkdownSize + 1); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	_, _, err = readMarkdown(path)
	if err == nil || !strings.Contains(err.Error(), "excede o limite") {
		t.Fatalf("readMarkdown() error = %v, want size limit error", err)
	}
}

func TestRunRequiresOneArgument(t *testing.T) {
	for _, args := range [][]string{nil, {"one.md", "two.md"}} {
		if err := Run(args); err != errUsage {
			t.Errorf("Run(%q) error = %v, want %v", args, err, errUsage)
		}
	}
}
