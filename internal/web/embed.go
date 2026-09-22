package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed page.html
var pageSource string

//go:embed styles.css
var styles string

//go:embed mermaid.min.js
var mermaid template.JS

type PageData struct {
	Title   string
	Path    string
	PDFName string
	Content template.HTML
	Token   string
}

func Page(data PageData) ([]byte, error) {
	tmpl, err := template.New("page").Parse(pageSource)
	if err != nil {
		return nil, fmt.Errorf("analisar template: %w", err)
	}

	view := struct {
		PageData
		Styles  template.CSS
		Mermaid template.JS
	}{data, template.CSS(styles), mermaid}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, view); err != nil {
		return nil, fmt.Errorf("executar template: %w", err)
	}
	return output.Bytes(), nil
}
