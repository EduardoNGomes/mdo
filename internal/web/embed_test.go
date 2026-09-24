package web

import (
	"html/template"
	"strings"
	"testing"

	"github.com/egomes/mdo/internal/theme"
)

func TestPageIncludesDocumentAndEmbeddedAssets(t *testing.T) {
	page, err := Page(PageData{
		Title:   `README <test>.md`,
		Path:    `/tmp/README <test>.md`,
		Content: template.HTML(`<h1 id="hello">Hello</h1>`),
		Token:   "abc123",
		Theme:   "tokyo-night-moon",
	})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	html := string(page)

	for _, expected := range []string{
		`README &lt;test&gt;.md`, `<h1 id="hello">Hello</h1>`,
		`<link rel="icon" href="data:,">`,
		`globalThis["mermaid"]`, `const token = "abc123"`,
		`fetch(`, `enableDiagramZoom`, `Zoom in`, `requestFullscreen`,
		`Diagram navigation`, `Pan diagram left`,
		`theme-select`, `prefers-color-scheme: light`, `rerenderMermaid`,
		`<wa-select`, `<wa-toast`, `<wa-button`, `themeToast.create`,
		`dataset.theme = "tokyo-night-moon"`, `Theme changed for this page only`, `mdo --theme`,
		`pdf-button`, `Download PDF`, `pdfBlob`, `window.mdoReady`, `/pdf`, `documentReady`,
		`mermaidSourceWithFallback`, `#59;`,
		`startViewTransition`, `to-light`, `to-dark`,
		`pre.mermaid`, `pre:not(.mermaid)`, `@media print`, `@page { size:A4`, `break-inside:avoid`,
		`link.getAttribute('href')`, `a[href^="#"]`, `target.scrollIntoView`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("Page() output does not contain %q", expected)
		}
	}
	for _, option := range theme.Options {
		if !strings.Contains(html, `value="`+option.ID+`"`) || !strings.Contains(html, option.Name) {
			t.Errorf("Page() missing theme option %q", option.Name)
		}
	}
	if strings.Contains(html, "localStorage") || strings.Contains(html, "sessionStorage") {
		t.Error("browser theme selection must not persist")
	}
	if strings.Contains(html, "window.alert(") {
		t.Error("theme changes must use the toast instead of a browser alert")
	}
	if strings.Contains(html, `test(link.href)`) {
		t.Error("Page() classifies resolved fragment URLs as external links")
	}
}
