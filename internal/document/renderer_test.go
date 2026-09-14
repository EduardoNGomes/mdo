package document

import (
	"strings"
	"testing"
)

func TestRenderMarkdownExtensions(t *testing.T) {
	source := []byte(`# Title

| Name | Value |
| --- | --- |
| mdo | reader |

- [x] done

~~removed~~

[^note]: footnote text

reference[^note]

Term
: Definition

` + "```go" + `
package main
` + "```" + `

` + "```mermaid" + `
flowchart LR
  A --> B
` + "```" + `
`)

	result, err := Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	html := string(result)
	for _, expected := range []string{
		`<h1 id="title">`, `<table>`, `type="checkbox"`, `<del>removed</del>`,
		`class="footnotes"`, `<dl>`, `class="chroma"`, `language-mermaid`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("Render() output does not contain %q", expected)
		}
	}
}

func TestRenderDoesNotAllowRawHTML(t *testing.T) {
	result, err := Render([]byte(`<script>alert("unsafe")</script>`))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(string(result), "<script>") {
		t.Fatalf("Render() preserved unsafe HTML: %s", result)
	}
}

func TestRenderBuildsHeaderFromYAMLFrontMatter(t *testing.T) {
	source := []byte("---\r\nkicker: Engineering\r\ntitle: Document title\r\nsubtitle: Technical refinement\r\naccent: orange\r\nbadge: Internal\r\ncobrand: smiles\r\nfooter_left: BRQ\r\nfooter_right: Confidential\r\nmeta:\r\n  Version: 1.0\r\n  Release: R10\r\n---\r\n\r\n## Content\r\n")

	result, err := Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	html := string(result)
	for _, expected := range []string{
		`class="document-header document-accent-orange"`,
		`<p class="document-kicker">Engineering</p>`,
		`<h1 id="document-title">Document title</h1>`,
		`<p class="document-subtitle">Technical refinement</p>`,
		`<span class="document-badge">Internal</span>`,
		`<span class="document-cobrand">smiles</span>`,
		`<dt>Version</dt><dd>1.0</dd>`,
		`<span>BRQ</span>`, `<span>Confidential</span>`,
		`<h2 id="content">Content</h2>`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("Render() output does not contain %q: %s", expected, html)
		}
	}
	if strings.Contains(html, "title: Document title") || strings.Contains(html, "meta:") {
		t.Fatalf("Render() exposed raw front matter: %s", html)
	}
}

func TestRenderEscapesFrontMatterValues(t *testing.T) {
	result, err := Render([]byte("---\ntitle: <script>alert('unsafe')</script>\n---\n"))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(string(result), "<script>") {
		t.Fatalf("Render() preserved unsafe front matter HTML: %s", result)
	}
}

func TestRenderKeepsUnclosedFrontMatter(t *testing.T) {
	result, err := Render([]byte("---\ntitle: Still visible"))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(string(result), "Still visible") {
		t.Fatalf("Render() silently discarded malformed front matter: %s", result)
	}
}
