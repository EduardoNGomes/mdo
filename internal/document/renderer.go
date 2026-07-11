package document

import (
	"bytes"
	"fmt"
	"html/template"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
		extension.DefinitionList,
		highlighting.NewHighlighting(
			highlighting.WithStyle("dracula"),
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

func Render(source []byte) (template.HTML, error) {
	var output bytes.Buffer
	if err := markdown.Convert(source, &output); err != nil {
		return "", fmt.Errorf("converter documento: %w", err)
	}
	return template.HTML(output.String()), nil
}
