package components

import (
	"context"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/a-h/templ"
)

const highlightMarkClass = "bg-retro-blush text-retro-ink px-1 border border-retro-ink"

func HighlightText(text, query string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, HighlightHTML(text, query))
		return err
	})
}

func HighlightHTML(text, query string) string {
	text = strings.TrimSpace(text)
	query = strings.TrimSpace(query)
	if text == "" {
		return ""
	}
	if query == "" {
		return html.EscapeString(text)
	}
	re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(query))
	if err != nil {
		return html.EscapeString(text)
	}
	indices := re.FindAllStringIndex(text, -1)
	if len(indices) == 0 {
		return html.EscapeString(text)
	}

	var b strings.Builder
	last := 0
	for _, idx := range indices {
		start, end := idx[0], idx[1]
		if start < last {
			continue
		}
		b.WriteString(html.EscapeString(text[last:start]))
		b.WriteString(`<mark class="`)
		b.WriteString(highlightMarkClass)
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(text[start:end]))
		b.WriteString(`</mark>`)
		last = end
	}
	b.WriteString(html.EscapeString(text[last:]))
	return b.String()
}
