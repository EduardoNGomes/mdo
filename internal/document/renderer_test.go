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
