package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePDF(t *testing.T) {
	if !chromeAvailable() {
		t.Skip("no Chrome-compatible browser is installed")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><h1>Documento</h1><h2>Seção</h2><script>window.mdoReady = true</script>`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), pdfTimeout)
	defer cancel()
	pdf, err := generatePDF(ctx, server.URL)
	if err != nil {
		t.Fatalf("generatePDF() error = %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("generatePDF() returned %q, want a PDF", pdf[:min(len(pdf), 8)])
	}
	if !bytes.Contains(pdf, []byte("/Outlines")) {
		t.Error("generatePDF() did not embed a document outline")
	}
}

func chromeAvailable() bool {
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "brave", "brave-browser", "msedge"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

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
	for _, args := range [][]string{nil, {"one.md", "two.md"}, {"--unknown", "one.md"}} {
		if err := Run(args); err != errUsage {
			t.Errorf("Run(%q) error = %v, want %v", args, err, errUsage)
		}
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want runOptions
		fail bool
	}{
		{name: "local", args: []string{"file.md"}, want: runOptions{markdownPath: "file.md"}},
		{name: "short live flag", args: []string{"-l", "file.md"}, want: runOptions{markdownPath: "file.md", live: true}},
		{name: "long live flag after file", args: []string{"file.md", "--live"}, want: runOptions{markdownPath: "file.md", live: true}},
		{name: "version", args: []string{"--version"}, want: runOptions{version: true}},
		{name: "short version", args: []string{"-v"}, want: runOptions{version: true}},
		{name: "missing file", args: []string{"--live"}, fail: true},
		{name: "version with file", args: []string{"--version", "file.md"}, fail: true},
		{name: "unknown flag", args: []string{"--share", "file.md"}, fail: true},
		{name: "two files", args: []string{"one.md", "two.md"}, fail: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if tt.fail {
				if err != errUsage {
					t.Fatalf("parseArgs(%q) error = %v, want %v", tt.args, err, errUsage)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("parseArgs(%q) = %#v, %v; want %#v, nil", tt.args, got, err, tt.want)
			}
		})
	}
}
