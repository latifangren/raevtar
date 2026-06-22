package service

import (
	"fmt"
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
	}
}

func (s *SiteMetaService) HomeSEO() model.SEOData {
	seo := s.DefaultSEO("/")
	seo.Description = "Raevtar — personal blog, server status board, docs, projects, and automation hooks running on postmarketOS."
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
