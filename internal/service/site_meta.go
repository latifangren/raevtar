package service

import (
	"fmt"
	"html"
	"strings"
	"time"

	"raevtar/internal/model"
)

const defaultSEODescription = "Personal platform for project notes, server monitoring, and automation."

type SitemapEntry struct {
	URL     string
	LastMod time.Time
}

type SiteMetaService struct {
	blog     *BlogService
	projects *ProjectService
	domain   string
}

func NewSiteMetaService(blog *BlogService, projects *ProjectService, domain string) *SiteMetaService {
	return &SiteMetaService{blog: blog, projects: projects, domain: strings.TrimSpace(domain)}
}

func (s *SiteMetaService) Domain() string {
	return s.domain
}

func (s *SiteMetaService) CanonicalURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://" + s.domain + path
}

func (s *SiteMetaService) DefaultSEO(path string) model.SEOData {
	canonicalPath := path
	if canonicalPath == "/lab/docs" {
		canonicalPath = "/docs"
	}
	return model.SEOData{
		Description:  defaultSEODescription,
		CanonicalURL: s.CanonicalURL(canonicalPath),
		SiteDomain:   s.domain,
	}
}

func (s *SiteMetaService) HomeSEO() model.SEOData {
	seo := s.DefaultSEO("/")
	seo.Description = "Raevtar — personal blog, server status board, docs, projects, and automation hooks."
	seo.JSONLD = model.MustJSONLD(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "WebSite",
		"name":        "Raevtar",
		"url":         seo.CanonicalURL,
		"description": seo.Description,
	})
	return seo
}

func (s *SiteMetaService) BlogPostSEO(post *model.Post) model.SEOData {
	seo := s.DefaultSEO("/blog/" + post.Slug)
	seo.ImageURL = s.CanonicalURL("/og-image/blog/" + post.Slug)
	description := strings.TrimSpace(post.Excerpt)
	if description == "" {
		description = defaultSEODescription
	}
	seo.Description = description
	payload := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "BlogPosting",
		"headline":      post.Title,
		"url":           seo.CanonicalURL,
		"description":   description,
		"datePublished": post.CreatedAt.UTC().Format(time.RFC3339),
		"dateModified":  maxTime(post.UpdatedAt, post.CreatedAt).UTC().Format(time.RFC3339),
	}
	if post.CategoryName != "" {
		payload["articleSection"] = post.CategoryName
	}
	if len(post.Tags) > 0 {
		payload["keywords"] = tagNames(post.Tags)
	}
	if strings.TrimSpace(post.CoverImageURL) != "" {
		payload["image"] = absoluteAssetURL(s.domain, post.CoverImageURL)
	}
	seo.JSONLD = model.MustJSONLD(payload)
	return seo
}

func (s *SiteMetaService) SitemapEntries() ([]SitemapEntry, error) {
	entries := []SitemapEntry{
		{URL: s.CanonicalURL("/")},
		{URL: s.CanonicalURL("/about")},
		{URL: s.CanonicalURL("/blog")},
		{URL: s.CanonicalURL("/contact")},
		{URL: s.CanonicalURL("/lab")},
		{URL: s.CanonicalURL("/docs")},
		{URL: s.CanonicalURL("/projects")},
		{URL: s.CanonicalURL("/topics")},
	}
	posts, _, err := s.blog.ListPosts("", 1, 500)
	if err != nil {
		return nil, fmt.Errorf("list posts for sitemap: %w", err)
	}
	for _, post := range posts {
		entries = append(entries, SitemapEntry{URL: s.CanonicalURL("/blog/" + post.Slug), LastMod: maxTime(post.UpdatedAt, post.CreatedAt)})
	}
	projects, _, err := s.projects.ListProjects(1, 500, ProjectListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list projects for sitemap: %w", err)
	}
	for _, project := range projects {
		lastMod := maxTime(project.UpdatedAt, project.CreatedAt)
		entries = append(entries, SitemapEntry{URL: s.CanonicalURL("/projects/" + project.Slug), LastMod: lastMod})
		entries = append(entries, SitemapEntry{URL: s.CanonicalURL("/projects/" + project.Slug + "/changelog"), LastMod: lastMod})
	}
	return entries, nil
}

func (s *SiteMetaService) ProjectSEO(project *model.Project) model.SEOData {
	seo := s.DefaultSEO("/projects/" + project.Slug)
	seo.ImageURL = s.CanonicalURL("/og-image/project/" + project.Slug)
	description := strings.TrimSpace(project.Excerpt)
	if description == "" {
		description = defaultSEODescription
	}
	seo.Description = description
	payload := map[string]any{
		"@context":      "https://schema.org",
		"@type":         "CreativeWork",
		"headline":      project.Title,
		"url":           seo.CanonicalURL,
		"description":   description,
		"datePublished": project.CreatedAt.UTC().Format(time.RFC3339),
		"dateModified":  maxTime(project.UpdatedAt, project.CreatedAt).UTC().Format(time.RFC3339),
	}
	if strings.TrimSpace(project.CoverImageURL) != "" {
		payload["image"] = absoluteAssetURL(s.domain, project.CoverImageURL)
	}
	seo.JSONLD = model.MustJSONLD(payload)
	return seo
}

func (s *SiteMetaService) LLMSText() (string, error) {
	posts, _, err := s.blog.ListPosts("", 1, 10)
	if err != nil {
		return "", fmt.Errorf("list posts for llms: %w", err)
	}
	projects, _, err := s.projects.ListProjects(1, 10, ProjectListOptions{})
	if err != nil {
		return "", fmt.Errorf("list projects for llms: %w", err)
	}
	lines := []string{
		"# Raevtar",
		"",
		"Raevtar is a public personal platform for project notes, server monitoring, docs, and automation.",
		"",
		"## Core Pages",
		"- Home: " + s.CanonicalURL("/"),
		"- About: " + s.CanonicalURL("/about"),
		"- Blog: " + s.CanonicalURL("/blog"),
		"- Projects: " + s.CanonicalURL("/projects"),
		"- Topics: " + s.CanonicalURL("/topics"),
		"- Docs: " + s.CanonicalURL("/docs"),
		"- Status: " + s.CanonicalURL("/dashboard"),
		"",
		"## Machine-readable Discovery",
		"- RSS: " + s.CanonicalURL("/blog/feed.xml"),
		"- Sitemap: " + s.CanonicalURL("/sitemap.xml"),
		"- Robots: " + s.CanonicalURL("/robots.txt"),
	}
	if len(posts) > 0 {
		lines = append(lines, "", "## Recent Blog Posts")
		for _, post := range posts {
			line := "- " + post.Title + ": " + s.CanonicalURL("/blog/"+post.Slug)
			if excerpt := strings.TrimSpace(post.Excerpt); excerpt != "" {
				line += " — " + excerpt
			}
			lines = append(lines, line)
		}
	}
	if len(projects) > 0 {
		lines = append(lines, "", "## Public Projects")
		for _, project := range projects {
			line := "- " + project.Title + ": " + s.CanonicalURL("/projects/"+project.Slug)
			if excerpt := strings.TrimSpace(project.Excerpt); excerpt != "" {
				line += " — " + excerpt
			}
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func (s *SiteMetaService) RSSFeed() (string, error) {
	posts, _, err := s.blog.ListPosts("", 1, 20)
	if err != nil {
		return "", fmt.Errorf("list posts for rss: %w", err)
	}

	domain := s.domain
	if domain == "" {
		domain = "raevtar.tech"
	}
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
`, XMLEscape(p.Title), domain, p.Slug, domain, p.Slug, XMLEscape(p.Excerpt), pubDate, p.CategorySlug)
	}

	xml += `</channel>
</rss>`

	return xml, nil
}

func (s *SiteMetaService) SitemapXML() (string, error) {
	entries, err := s.SitemapEntries()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, entry := range entries {
		b.WriteString("  <url>\n")
		b.WriteString("    <loc>" + XMLEscape(entry.URL) + "</loc>\n")
		if !entry.LastMod.IsZero() {
			b.WriteString("    <lastmod>" + entry.LastMod.UTC().Format("2006-01-02T15:04:05Z") + "</lastmod>\n")
		}
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>")
	return b.String(), nil
}

func (s *SiteMetaService) RobotsTxt() string {
	domain := s.domain
	if domain == "" {
		domain = "example.com"
	}
	return fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: https://%s/sitemap.xml\n", domain)
}

func (s *SiteMetaService) OGImageBlogSVG(slug string) (string, error) {
	post, err := s.blog.GetPublishedPost(slug)
	if err != nil {
		return "", err
	}
	return s.OGImageSVG(post.Title, "Blog dispatch from Raevtar"), nil
}

func (s *SiteMetaService) OGImageProjectSVG(slug string) (string, error) {
	project, err := s.projects.GetPublishedProject(slug)
	if err != nil {
		return "", err
	}
	return s.OGImageSVG(project.Title, "Project from small-machine lab"), nil
}

func (s *SiteMetaService) OGImageSVG(title, subtitle string) string {
	domain := s.domain
	if domain == "" {
		domain = "raevtar.tech"
	}
	return fmt.Sprintf(`<svg width="1200" height="630" viewBox="0 0 1200 630" xmlns="http://www.w3.org/2000/svg">
		<rect width="1200" height="630" fill="#F5F2ED"/>
		<!-- Layered borders -->
		<rect x="20" y="20" width="1160" height="590" fill="none" stroke="#2D3748" stroke-width="8"/>
		<rect x="40" y="40" width="1160" height="590" fill="none" stroke="#2D3748" stroke-width="8"/>
		<!-- Background highlight -->
		<rect x="30" y="30" width="1140" height="570" fill="#FACC15" stroke="#2D3748" stroke-width="4"/>
		
		<!-- Subtitle -->
		<text x="60" y="100" font-family="sans-serif" font-size="24" font-weight="900" fill="#2D3748" text-transform="uppercase">%s</text>
		
		<!-- Main Title (wrapped manually for simplicity in SVG) -->
		<text x="60" y="300" font-family="sans-serif" font-size="72" font-weight="900" fill="#2D3748">%s</text>
		
		<!-- Branding -->
		<rect x="60" y="500" width="200" height="60" fill="#2D3748"/>
		<text x="80" y="540" font-family="sans-serif" font-size="32" font-weight="bold" fill="#F5F2ED">`+domain+`</text>
	</svg>`, html.EscapeString(subtitle), html.EscapeString(title))
}

func XMLEscape(s string) string {
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

func absoluteAssetURL(domain, asset string) string {
	if strings.HasPrefix(asset, "http://") || strings.HasPrefix(asset, "https://") {
		return asset
	}
	if !strings.HasPrefix(asset, "/") {
		asset = "/" + asset
	}
	return "https://" + domain + asset
}

func tagNames(tags []model.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		name := strings.TrimSpace(tag.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

func maxTime(values ...time.Time) time.Time {
	var latest time.Time
	for _, value := range values {
		if value.After(latest) {
			latest = value
		}
	}
	return latest
}
