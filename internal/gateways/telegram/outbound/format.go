package outbound

import (
	"html"
	"regexp"
	"strings"
)

var markdownItalic = regexp.MustCompile(`\*([^*\n]+)\*`)

func formatForTelegramHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var b strings.Builder

	last := 0
	for _, loc := range markdownItalic.FindAllStringIndex(raw, -1) {
		b.WriteString(html.EscapeString(raw[last:loc[0]]))
		b.WriteString("<i>")
		b.WriteString(html.EscapeString(raw[loc[0]+1 : loc[1]-1]))
		b.WriteString("</i>")

		last = loc[1]
	}

	b.WriteString(html.EscapeString(raw[last:]))

	return b.String()
}
