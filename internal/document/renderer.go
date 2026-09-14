package document

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type metadataItem struct {
	Label string
	Value string
}

type frontMatter struct {
	Kicker      string
	Title       string
	Subtitle    string
	Badge       string
	Cobrand     string
	AccentClass string
	FooterLeft  string
	FooterRight string
	Meta        []metadataItem
}

var frontMatterHeader = template.Must(template.New("front-matter").Parse(`{{if .Title}}<header class="document-header {{.AccentClass}}">
{{if .Kicker}}<p class="document-kicker">{{.Kicker}}</p>{{end}}
<h1 id="document-title">{{.Title}}</h1>
{{if .Subtitle}}<p class="document-subtitle">{{.Subtitle}}</p>{{end}}
{{if or .Badge .Cobrand}}<div class="document-badges">{{if .Badge}}<span class="document-badge">{{.Badge}}</span>{{end}}{{if .Cobrand}}<span class="document-cobrand">{{.Cobrand}}</span>{{end}}</div>{{end}}
{{if .Meta}}<dl class="document-meta">{{range .Meta}}<div><dt>{{.Label}}</dt><dd>{{.Value}}</dd></div>{{end}}</dl>{{end}}
{{if or .FooterLeft .FooterRight}}<div class="document-header-footer">{{if .FooterLeft}}<span>{{.FooterLeft}}</span>{{end}}{{if .FooterRight}}<span>{{.FooterRight}}</span>{{end}}</div>{{end}}
</header>{{end}}`))

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
	metadata, source, hasFrontMatter := parseFrontMatter(source)

	var output bytes.Buffer
	if hasFrontMatter {
		if err := frontMatterHeader.Execute(&output, metadata); err != nil {
			return "", fmt.Errorf("renderizar front matter: %w", err)
		}
	}
	if err := markdown.Convert(source, &output); err != nil {
		return "", fmt.Errorf("converter documento: %w", err)
	}
	return template.HTML(output.String()), nil
}

// parseFrontMatter reads the simple key/value YAML front matter commonly used
// by Markdown documents. It also preserves the order of entries nested below
// "meta", which is useful for a predictable document header.
func parseFrontMatter(source []byte) (frontMatter, []byte, bool) {
	var metadata frontMatter
	lineEnd := bytes.IndexByte(source, '\n')
	if lineEnd < 0 || !isFrontMatterDelimiter(source[:lineEnd]) {
		return metadata, source, false
	}

	blockStart := lineEnd + 1
	for offset := lineEnd + 1; offset < len(source); {
		nextLineEnd := bytes.IndexByte(source[offset:], '\n')
		if nextLineEnd < 0 {
			if isFrontMatterDelimiter(source[offset:]) {
				parseFrontMatterBlock(source[blockStart:offset], &metadata)
				return metadata, source[len(source):], true
			}
			break
		}

		nextLineEnd += offset
		if isFrontMatterDelimiter(source[offset:nextLineEnd]) {
			parseFrontMatterBlock(source[blockStart:offset], &metadata)
			return metadata, source[nextLineEnd+1:], true
		}
		offset = nextLineEnd + 1
	}

	// Keep malformed/unclosed front matter visible instead of silently dropping
	// the entire document.
	return metadata, source, false
}

func parseFrontMatterBlock(block []byte, metadata *frontMatter) {
	section := ""
	for _, rawLine := range bytes.Split(block, []byte{'\n'}) {
		line := strings.TrimSuffix(string(rawLine), "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		separator := strings.IndexByte(trimmed, ':')
		if separator < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:separator])
		value := unquoteFrontMatterValue(strings.TrimSpace(trimmed[separator+1:]))
		indented := len(line) > len(strings.TrimLeft(line, " \t"))

		if indented && section == "meta" {
			metadata.Meta = append(metadata.Meta, metadataItem{Label: key, Value: value})
			continue
		}

		section = ""
		if value == "" {
			section = strings.ToLower(key)
			continue
		}
		switch strings.ToLower(key) {
		case "kicker":
			metadata.Kicker = value
		case "title":
			metadata.Title = value
		case "subtitle":
			metadata.Subtitle = value
		case "badge":
			metadata.Badge = value
		case "cobrand":
			metadata.Cobrand = value
		case "accent":
			metadata.AccentClass = frontMatterAccentClass(value)
		case "footer_left":
			metadata.FooterLeft = value
		case "footer_right":
			metadata.FooterRight = value
		}
	}
}

func unquoteFrontMatterValue(value string) string {
	if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') ||
		(value[0] == '"' && value[len(value)-1] == '"')) {
		return value[1 : len(value)-1]
	}
	return value
}

func frontMatterAccentClass(accent string) string {
	switch strings.ToLower(strings.TrimSpace(accent)) {
	case "laranja", "orange":
		return "document-accent-orange"
	case "azul", "blue":
		return "document-accent-blue"
	case "roxo", "purple":
		return "document-accent-purple"
	case "vermelho", "red":
		return "document-accent-red"
	default:
		return ""
	}
}

func isFrontMatterDelimiter(line []byte) bool {
	line = bytes.TrimSuffix(line, []byte{'\r'})
	return bytes.Equal(line, []byte("---"))
}
