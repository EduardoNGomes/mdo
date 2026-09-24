package app

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/egomes/mdo/internal/theme"
	"github.com/egomes/mdo/internal/web"
)

func TestGeneratePDF(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("headless Chrome sandbox is unavailable in CI")
	}
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
		{name: "theme", args: []string{"--theme"}, want: runOptions{theme: true}},
		{name: "short version", args: []string{"-v"}, want: runOptions{version: true}},
		{name: "missing file", args: []string{"--live"}, fail: true},
		{name: "version with file", args: []string{"--version", "file.md"}, fail: true},
		{name: "theme with file", args: []string{"--theme", "file.md"}, fail: true},
		{name: "theme with live", args: []string{"--theme", "--live"}, fail: true},
		{name: "theme with version", args: []string{"--theme", "--version"}, fail: true},
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

func TestRunThemeSelectsWithoutOpeningBrowser(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	previous := selectTheme
	t.Cleanup(func() { selectTheme = previous })
	selectTheme = func(current string) (string, error) {
		if current != "" {
			t.Errorf("initial theme = %q, want unset", current)
		}
		return "catppuccin-macchiato", nil
	}
	if err := Run([]string{"--theme"}); err != nil {
		t.Fatal(err)
	}
	path, err := theme.Path()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := theme.Load(path); err != nil || got != "catppuccin-macchiato" {
		t.Fatalf("saved theme = %q, %v", got, err)
	}
	selectTheme = func(current string) (string, error) {
		if current != "catppuccin-macchiato" {
			t.Errorf("selector current = %q", current)
		}
		return "", huh.ErrUserAborted
	}
	if err := Run([]string{"--theme"}); err != nil {
		t.Fatal(err)
	}
	if got, err := theme.Load(path); err != nil || got != "catppuccin-macchiato" {
		t.Fatalf("theme after cancellation = %q, %v", got, err)
	}
}

func TestBrowserThemeSelection(t *testing.T) {
	if os.Getenv("CI") != "" || !chromeAvailable() {
		t.Skip("headless Chrome is unavailable")
	}
	pageHTML, err := web.Page(web.PageData{
		Title: "Themes", Path: "/tmp/themes.md", Token: "test-token",
		Theme: "catppuccin-latte", Content: template.HTML("<h1>Themes</h1><table><tr><td><code id=inline-code>default</code></td></tr></table><div class=chroma><span class=k>func</span><span class=s>string</span><span class=c>comment</span></div><pre><code class=language-mermaid>flowchart LR\nA-->B</code></pre>"),
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(pageHTML)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	browserCtx, cancelBrowser := chromedp.NewContext(ctx)
	defer cancelBrowser()
	var requestMu sync.Mutex
	var externalRequests []string
	chromedp.ListenTarget(browserCtx, func(event any) {
		if request, ok := event.(*network.EventRequestWillBeSent); ok {
			url := request.Request.URL
			if !strings.HasPrefix(url, server.URL) && !strings.HasPrefix(url, "data:") {
				requestMu.Lock()
				externalRequests = append(externalRequests, url)
				requestMu.Unlock()
			}
		}
	})
	var initial, changed struct {
		Theme     string `json:"theme"`
		Selection string `json:"selection"`
		Alert     string `json:"alert"`
		Color     string `json:"color"`
		Count     int    `json:"count"`
		Diagrams  int    `json:"diagrams"`
	}
	var palettes []struct {
		ID          string `json:"id"`
		Background  string `json:"background"`
		Syntax      string `json:"syntax"`
		Inline      string `json:"inline"`
		String      string `json:"string"`
		Comment     string `json:"comment"`
		Control     string `json:"control"`
		WantControl string `json:"wantControl"`
	}
	if err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL),
		chromedp.WaitReady(".theme-select"),
		chromedp.Poll(`window.mdoReady === true`, nil),
		chromedp.Evaluate(`(() => ({theme:document.documentElement.dataset.theme, selection:document.querySelector('.theme-select').value, color:getComputedStyle(document.documentElement).getPropertyValue('--bg').trim(), count:document.querySelectorAll('.theme-select wa-option').length, diagrams:document.querySelectorAll('.mermaid svg').length}))()`, &initial),
		chromedp.Evaluate(`(() => {
          const root = document.documentElement;
          // Sample settled palette colors, not the component's transition frames.
          const combobox = document.querySelector('.theme-select').shadowRoot.querySelector('[part~=combobox]');
          combobox.style.transition = 'none';
          const hex = (selector, property = 'color') => '#' + getComputedStyle(typeof selector === 'string' ? document.querySelector(selector) : selector)[property].match(/\d+/g).slice(0, 3).map(n => Number(n).toString(16).padStart(2, '0')).join('');
          return [...document.querySelectorAll('.theme-select wa-option')].map(option => {
            root.dataset.theme = option.value;
            root.dataset.appearance = ['default-light','catppuccin-latte','tokyo-night-day','alucard-classic'].includes(option.value) ? 'light' : 'dark';
            return {id:option.value, background:getComputedStyle(root).getPropertyValue('--bg').trim(),
              control:hex(document.querySelector('.theme-select').shadowRoot.querySelector('[part~=combobox]'), 'backgroundColor'), wantControl:getComputedStyle(root).getPropertyValue('--surface-strong').trim(), syntax:hex('.chroma .k'), inline:hex('#inline-code'), string:hex('.chroma .s'), comment:hex('.chroma .c')};
          });
        })()`, &palettes),
		chromedp.Evaluate(`document.querySelector('.theme-select').value = 'dracula-classic'; document.querySelector('.theme-select').dispatchEvent(new Event('change', {bubbles:true}))`, nil),
		chromedp.Poll(`document.querySelector('.theme-toast wa-toast-item') && document.querySelector('.theme-toast').matches(':popover-open') && document.documentElement.dataset.theme === 'dracula-classic'`, nil),
		chromedp.Evaluate(`(() => ({theme:document.documentElement.dataset.theme, alert:document.querySelector('.theme-toast wa-toast-item').textContent, color:getComputedStyle(document.documentElement).getPropertyValue('--bg').trim(), diagrams:document.querySelectorAll('.mermaid svg').length}))()`, &changed),
	); err != nil {
		t.Fatal(err)
	}
	if initial.Theme != "catppuccin-latte" || initial.Selection != "catppuccin-latte" || initial.Count != len(theme.Options) || initial.Color != "#eff1f5" || initial.Diagrams != 1 {
		t.Errorf("initial browser theme = %+v", initial)
	}
	if changed.Theme != "dracula-classic" || changed.Color != "#282a36" || changed.Diagrams != 1 || !strings.Contains(changed.Alert, "mdo --theme") {
		t.Errorf("changed browser theme = %+v", changed)
	}
	// Expected values from the published palette and syntax references in styles.css.
	// Columns: background, inline Markdown code, keywords, strings, comments.
	expected := map[string][5]string{
		"default":              {"#020618", "#9ff5cb", "#ff79c6", "#f1fa8c", "#6272a4"},
		"default-light":        {"#f4f8f6", "#006d40", "#ff79c6", "#f1fa8c", "#6272a4"},
		"catppuccin-latte":     {"#eff1f5", "#40a02b", "#8839ef", "#40a02b", "#7c7f93"},
		"catppuccin-frappe":    {"#303446", "#a6d189", "#ca9ee6", "#a6d189", "#949cbb"},
		"catppuccin-macchiato": {"#24273a", "#a6da95", "#c6a0f6", "#a6da95", "#939ab7"},
		"catppuccin-mocha":     {"#1e1e2e", "#a6e3a1", "#cba6f7", "#a6e3a1", "#9399b2"},
		"nord":                 {"#2e3440", "#a3be8c", "#81a1c1", "#a3be8c", "#4c566a"},
		"tokyo-night":          {"#1a1b26", "#7aa2f7", "#bb9af7", "#9ece6a", "#565f89"},
		"tokyo-night-storm":    {"#24283b", "#7aa2f7", "#bb9af7", "#9ece6a", "#565f89"},
		"tokyo-night-moon":     {"#222436", "#82aaff", "#c099ff", "#c3e88d", "#636da6"},
		"tokyo-night-day":      {"#e1e2e7", "#2e7de9", "#9854f1", "#587539", "#848cb5"},
		"dracula-classic":      {"#282a36", "#50fa7b", "#ff79c6", "#f1fa8c", "#6272a4"},
		"alucard-classic":      {"#fffbeb", "#14710a", "#a3144d", "#846e15", "#6c664b"},
	}
	if len(palettes) != len(expected) {
		t.Errorf("got %d browser palettes, want %d", len(palettes), len(expected))
	}
	for _, palette := range palettes {
		if palette.Control != palette.WantControl {
			t.Errorf("%s dropdown background = %s, want %s", palette.ID, palette.Control, palette.WantControl)
		}
		got := [5]string{palette.Background, palette.Inline, palette.Syntax, palette.String, palette.Comment}
		if want, ok := expected[palette.ID]; !ok || got != want {
			t.Errorf("%s rendered colors = %v, want %v", palette.ID, got, want)
		}
	}
	var reloaded string
	if err := chromedp.Run(browserCtx,
		chromedp.Navigate(server.URL),
		chromedp.WaitReady(".theme-select"),
		chromedp.Evaluate(`document.documentElement.dataset.theme`, &reloaded),
	); err != nil {
		t.Fatal(err)
	}
	if reloaded != "catppuccin-latte" {
		t.Errorf("theme after reload = %q, want saved CLI choice", reloaded)
	}
	pdfContext, cancelPDF := context.WithTimeout(context.Background(), pdfTimeout)
	defer cancelPDF()
	pdf, err := generatePDF(pdfContext, server.URL+"?pdf=1")
	if err != nil || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("themed PDF = %d bytes, %v", len(pdf), err)
	}
	// The one-shot server normally closes once the page is prepared. Components
	// must remain fully interactive without fetching JavaScript, CSS, or icons.
	server.Close()
	if err := chromedp.Run(browserCtx,
		chromedp.Poll(`window.mdoReady === true`, nil),
		chromedp.Evaluate(`document.querySelector('.theme-select').focus()`, nil),
		chromedp.KeyEvent(kb.ArrowDown),
		chromedp.KeyEvent(kb.End),
		chromedp.KeyEvent(kb.Enter),
		chromedp.Poll(`document.documentElement.dataset.theme === 'alucard-classic' && !!document.querySelector('.theme-toast wa-toast-item')`, nil),
		chromedp.Evaluate(`document.querySelector('.theme-toast wa-toast-item').shadowRoot.querySelector('[part="close-button"]').click()`, nil),
		chromedp.Poll(`document.querySelector('.theme-toast wa-toast-item') === null`, nil),
	); err != nil {
		t.Fatalf("offline keyboard selection or toast dismissal: %v", err)
	}
	requestMu.Lock()
	defer requestMu.Unlock()
	if len(externalRequests) > 0 {
		t.Errorf("components requested external assets: %v", externalRequests)
	}
}
