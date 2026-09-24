package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

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

	var output bytes.Buffer
	if err := tmpl.Execute(&output, view); err != nil {
		return nil, fmt.Errorf("executar template: %w", err)
	}
	return output.Bytes(), nil
}
