package handler

import (
	"fmt"
	"net/http"
	"time"
)

// RSSFeed generates an RSS 2.0 feed of recent blog posts.
func (h *Handler) rssFeed(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.svc.Blog.ListPosts("", 1, 20)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	domain := h.cfg.Domain
	now := time.Now().Format(time.RFC1123Z)

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:content="http://purl.org/rss/1.0/modules/content/">
<channel>
	<title>Raevtar</title>
	<link>https://` + domain + `</link>
	<description>Personal blog about development, AI agents, security, and DevOps</description>
	<language>en</language>
	<lastBuildDate>` + now + `</lastBuildDate>
	<atom:link href="https://` + domain + `/blog/feed.xml" rel="self" type="application/rss+xml"/>
`

	for _, p := range posts {
		pubDate := p.CreatedAt.Format(time.RFC1123Z)
		xml += fmt.Sprintf(`	<item>
		<title>%s</title>
		<link>https://%s/blog/%s</link>
		<guid isPermaLink="true">https://%s/blog/%s</guid>
		<description><![CDATA[%s]]></description>
		<pubDate>%s</pubDate>
		<category>%s</category>
	</item>
`, xmlEscape(p.Title), domain, p.Slug, domain, p.Slug, xmlEscape(p.Excerpt), pubDate, p.CategorySlug)
	}

	xml += `</channel>
</rss>`

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(xml))
}

func xmlEscape(s string) string {
	escaped := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '&':
			escaped = append(escaped, "&amp;"...)
		case '<':
			escaped = append(escaped, "&lt;"...)
		case '>':
			escaped = append(escaped, "&gt;"...)
		case '"':
			escaped = append(escaped, "&quot;"...)
		case '\'':
			escaped = append(escaped, "&apos;"...)
		default:
			escaped = append(escaped, c)
		}
	}
	return string(escaped)
}
