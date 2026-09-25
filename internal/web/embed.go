package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"regexp"

	"github.com/egomes/mdo/internal/theme"
)

//go:embed page.html
var pageSource string

//go:embed styles.css
var styles string

//go:embed mermaid.min.js
var mermaid template.JS

//go:embed components.min.js
var components template.JS

//go:embed components.min.css
var componentStyles string

var mermaidCode = regexp.MustCompile(`<code\s+class=(?:"language-mermaid"|'language-mermaid'|language-mermaid)[\s>]`)

type PageData struct {
	Title   string
	Path    string
	PDFName string
	Content template.HTML
	Token   string
	Theme   string
}

func Page(data PageData) ([]byte, error) {
	tmpl, err := template.New("page").Parse(pageSource)
	if err != nil {
		return nil, fmt.Errorf("analisar template: %w", err)
	}

	view := struct {
		PageData
		Styles          template.CSS
		Mermaid         template.JS
		Themes          []theme.Option
		Components      template.JS
		ComponentStyles template.CSS
	}{data, template.CSS(styles), mermaid, theme.Options, components, template.CSS(componentStyles)}
	// Goldmark emits this class for fenced Mermaid blocks. Inspect the rendered
	// content so mentions of Mermaid in prose or escaped code don't load the bundle.
	if !mermaidCode.MatchString(string(data.Content)) {
		view.Mermaid = ""
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, view); err != nil {
		return nil, fmt.Errorf("executar template: %w", err)
	}
	return output.Bytes(), nil
}
