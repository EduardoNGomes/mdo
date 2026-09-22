package web

import (
	"html/template"
	"strings"
	"testing"
)

func TestPageIncludesDocumentAndEmbeddedAssets(t *testing.T) {
	page, err := Page(PageData{
		Title:   `README <test>.md`,
		Path:    `/tmp/README <test>.md`,
		Content: template.HTML(`<h1 id="hello">Hello</h1>`),
		Token:   "abc123",
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
		`theme-button`, `prefers-color-scheme: light`, `rerenderMermaid`,
		`pdf-button`, `Download PDF`, `pdfBlob`, `window.mdoReady`, `/pdf`, `documentReady`,
		`mermaidSourceWithFallback`, `#59;`,
		`startViewTransition`, `to-light`, `to-dark`, `theme-icon-sun`,
		`pre.mermaid`, `pre:not(.mermaid)`, `@media print`, `@page { size:A4`, `break-inside:avoid`,
		`link.getAttribute('href')`, `a[href^="#"]`, `target.scrollIntoView`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("Page() output does not contain %q", expected)
		}
	}
	if strings.Contains(html, `test(link.href)`) {
		t.Error("Page() classifies resolved fragment URLs as external links")
	}
}
