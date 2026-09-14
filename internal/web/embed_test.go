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
		`globalThis["mermaid"]`, `const token = "abc123"`,
		`fetch(`, `enableDiagramZoom`, `Aumentar zoom`, `requestFullscreen`,
		`Navegação do diagrama`, `Mover diagrama para a esquerda`,
		`theme-button`, `prefers-color-scheme: light`, `rerenderMermaid`,
		`startViewTransition`, `to-light`, `to-dark`, `theme-icon-sun`,
		`pre.mermaid`, `pre:not(.mermaid)`,
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
