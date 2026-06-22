package service

import (
	"strings"
	"testing"

	"raevtar/internal/model"
)

func TestSiteMetaServiceCanonicalURL(t *testing.T) {
	state := newTestServices(t)

	tests := []struct {
		path string
		want string
	}{
		{path: "/", want: "https://raevtar.test/"},
		{path: "/blog", want: "https://raevtar.test/blog"},
		{path: "/blog/article-slug", want: "https://raevtar.test/blog/article-slug"},
		{path: "/projects", want: "https://raevtar.test/projects"},
		{path: "/status", want: "https://raevtar.test/status"},
		{path: "/about", want: "https://raevtar.test/about"},
		{path: "", want: "https://raevtar.test/"},
		{path: "no-leading-slash", want: "https://raevtar.test/no-leading-slash"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := state.svc.SiteMeta.CanonicalURL(tt.path)
			if got != tt.want {
				t.Fatalf("CanonicalURL(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestSiteMetaServiceDefaultSEO(t *testing.T) {
	state := newTestServices(t)

	seo := state.svc.SiteMeta.DefaultSEO("/blog")

	if seo.Description == "" {
		t.Fatalf("DefaultSEO description should not be empty")
	}
	if seo.CanonicalURL != "https://raevtar.test/blog" {
		t.Fatalf("DefaultSEO CanonicalURL = %q, want https://raevtar.test/blog", seo.CanonicalURL)
	}
	if seo.ImageURL != "" {
		t.Fatalf("DefaultSEO ImageURL = %q, want empty", seo.ImageURL)
	}
	if seo.JSONLD != "" {
		t.Fatalf("DefaultSEO JSONLD = %q, want empty", seo.JSONLD)
	}
}

func TestSiteMetaServiceHomeSEO(t *testing.T) {
	state := newTestServices(t)

	seo := state.svc.SiteMeta.HomeSEO()

	if seo.Description == "" {
		t.Fatalf("HomeSEO description should not be empty")
	}
	if seo.CanonicalURL != "https://raevtar.test/" {
		t.Fatalf("HomeSEO CanonicalURL = %q, want https://raevtar.test/", seo.CanonicalURL)
	}
	if seo.JSONLD == "" {
		t.Fatalf("HomeSEO JSONLD should not be empty")
	}
	if !strings.Contains(seo.JSONLD, "WebSite") {
		t.Fatalf("HomeSEO JSONLD = %q, want to contain WebSite", seo.JSONLD)
	}
	if seo.ImageURL != "" {
		t.Fatalf("HomeSEO ImageURL = %q, want empty", seo.ImageURL)
	}
}

func TestSiteMetaServiceBlogPostSEO(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Test Blog Post for SEO",
		ContentMD:    "# Test SEO Post\n\nContent here.",
		Excerpt:      "A test blog post excerpt for SEO validation.",
		Published:    true,
		Tags:         []string{"go", "testing"},
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	seo := state.svc.SiteMeta.BlogPostSEO(post)

	if seo.CanonicalURL != "https://raevtar.test/blog/test-blog-post-for-seo" {
		t.Fatalf("BlogPostSEO CanonicalURL = %q, want https://raevtar.test/blog/test-blog-post-for-seo", seo.CanonicalURL)
	}
	if seo.Description != "A test blog post excerpt for SEO validation." {
		t.Fatalf("BlogPostSEO Description = %q, want excerpt content", seo.Description)
	}
	if seo.JSONLD == "" {
		t.Fatalf("BlogPostSEO JSONLD should not be empty")
	}
	if !strings.Contains(seo.JSONLD, "BlogPosting") {
		t.Fatalf("BlogPostSEO JSONLD = %q, want to contain BlogPosting", seo.JSONLD)
	}
	if !strings.Contains(seo.JSONLD, `"headline":"Test Blog Post for SEO"`) {
		t.Fatalf("BlogPostSEO JSONLD = %q, want to contain headline", seo.JSONLD)
	}
}

func TestSiteMetaServiceBlogPostSEOWithCoverImage(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug:  "ai-agent",
		Title:         "SEO Post With Cover",
		ContentMD:     "# Cover Image\n\nTesting cover image in SEO.",
		Excerpt:       "Post with cover image for SEO.",
		CoverImageURL: "/uploads/test-cover.png",
		Published:     true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	seo := state.svc.SiteMeta.BlogPostSEO(post)

	if seo.ImageURL == "" {
		t.Fatalf("BlogPostSEO ImageURL should not be empty when cover image is set")
	}
	if seo.ImageURL != "https://raevtar.test/og-image/blog/seo-post-with-cover" {
		t.Fatalf("BlogPostSEO ImageURL = %q, want https://raevtar.test/og-image/blog/seo-post-with-cover", seo.ImageURL)
	}
}

func TestSiteMetaServiceBlogPostSEOWithoutCoverImage(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "security",
		Title:        "SEO Post Without Cover",
		ContentMD:    "# No Cover\n\nNo cover image for this post.",
		Excerpt:      "Post without cover image.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	seo := state.svc.SiteMeta.BlogPostSEO(post)

	if seo.ImageURL == "" {
		t.Fatalf("BlogPostSEO ImageURL should always be set to OG image URL")
	}
	if strings.Contains(seo.JSONLD, `"image"`) {
		t.Fatalf("BlogPostSEO JSONLD should not contain image key when no cover image is set: %s", seo.JSONLD)
	}
}

func TestSiteMetaServiceBlogPostSEOWithEmptyExcerpt(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Post Without Excerpt",
		ContentMD:    "# No Excerpt\n\nContent without excerpt.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	seo := state.svc.SiteMeta.BlogPostSEO(post)

	if seo.Description == "" {
		t.Fatalf("BlogPostSEO Description should fall back to default when excerpt is empty")
	}
	if seo.Description == "" {
		t.Fatalf("BlogPostSEO Description should not be empty")
	}
}

func TestSiteMetaServiceProjectSEO(t *testing.T) {
	state := newTestServices(t)

	project, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Test SEO Project",
		ContentMD: "# Test Project\n\nProject for SEO test.",
		Excerpt:   "A test project for SEO validation.",
		Published: true,
		State:     model.ProjectStateActive,
		Featured:  true,
		SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	seo := state.svc.SiteMeta.ProjectSEO(project)

	if seo.CanonicalURL != "https://raevtar.test/projects/test-seo-project" {
		t.Fatalf("ProjectSEO CanonicalURL = %q, want https://raevtar.test/projects/test-seo-project", seo.CanonicalURL)
	}
	if seo.Description != "A test project for SEO validation." {
		t.Fatalf("ProjectSEO Description = %q, want excerpt content", seo.Description)
	}
	if seo.JSONLD == "" {
		t.Fatalf("ProjectSEO JSONLD should not be empty")
	}
	if !strings.Contains(seo.JSONLD, "CreativeWork") {
		t.Fatalf("ProjectSEO JSONLD = %q, want to contain CreativeWork", seo.JSONLD)
	}
	if !strings.Contains(seo.JSONLD, `"headline":"Test SEO Project"`) {
		t.Fatalf("ProjectSEO JSONLD = %q, want to contain headline", seo.JSONLD)
	}
	if seo.ImageURL != "https://raevtar.test/og-image/project/test-seo-project" {
		t.Fatalf("ProjectSEO ImageURL = %q, want https://raevtar.test/og-image/project/test-seo-project", seo.ImageURL)
	}
}

func TestSiteMetaServiceProjectSEOWithEmptyExcerpt(t *testing.T) {
	state := newTestServices(t)

	project, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Project No Excerpt",
		ContentMD: "# No Excerpt Project\n\nContent.",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	seo := state.svc.SiteMeta.ProjectSEO(project)

	if seo.Description == "" {
		t.Fatalf("ProjectSEO Description should not be empty, should fall back to default")
	}
}

func TestSiteMetaServiceSitemapEntries(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Sitemap Test Post",
		ContentMD:    "# Sitemap Post\n\nContent.",
		Excerpt:      "Sitemap test post.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	_, err = state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Sitemap Test Project",
		ContentMD: "# Sitemap Project\n\nContent.",
		Excerpt:   "Sitemap test project.",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	entries, err := state.svc.SiteMeta.SitemapEntries()
	if err != nil {
		t.Fatalf("SitemapEntries: %v", err)
	}

	expectedPaths := map[string]bool{
		"https://raevtar.test/":                                        false,
		"https://raevtar.test/about":                                   false,
		"https://raevtar.test/blog":                                    false,
		"https://raevtar.test/contact":                                 false,
		"https://raevtar.test/lab":                                     false,
		"https://raevtar.test/docs":                                    false,
		"https://raevtar.test/projects":                                false,
		"https://raevtar.test/topics":                                  false,
		"https://raevtar.test/blog/sitemap-test-post":                  false,
		"https://raevtar.test/projects/sitemap-test-project":           false,
		"https://raevtar.test/projects/sitemap-test-project/changelog": false,
	}

	for _, entry := range entries {
		if _, ok := expectedPaths[entry.URL]; ok {
			expectedPaths[entry.URL] = true
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Fatalf("SitemapEntries missing expected path: %s", path)
		}
	}

	if len(entries) < len(expectedPaths) {
		t.Fatalf("SitemapEntries count = %d, want at least %d", len(entries), len(expectedPaths))
	}
}

func TestSiteMetaServiceSitemapEntriesNoContent(t *testing.T) {
	state := newTestServices(t)

	entries, err := state.svc.SiteMeta.SitemapEntries()
	if err != nil {
		t.Fatalf("SitemapEntries: %v", err)
	}

	expectedStaticPaths := map[string]bool{
		"https://raevtar.test/":         false,
		"https://raevtar.test/about":    false,
		"https://raevtar.test/blog":     false,
		"https://raevtar.test/contact":  false,
		"https://raevtar.test/lab":      false,
		"https://raevtar.test/docs":     false,
		"https://raevtar.test/projects": false,
		"https://raevtar.test/topics":   false,
	}

	for _, entry := range entries {
		if _, ok := expectedStaticPaths[entry.URL]; ok {
			expectedStaticPaths[entry.URL] = true
		}
	}

	for path, found := range expectedStaticPaths {
		if !found {
			t.Fatalf("SitemapEntries missing static path: %s", path)
		}
	}

	if len(entries) != len(expectedStaticPaths) {
		t.Fatalf("SitemapEntries count = %d, want %d (no posts or projects should be present)", len(entries), len(expectedStaticPaths))
	}
}

func TestSiteMetaServiceLLMSText(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "kernel-embedded",
		Title:        "LLMS Kernel Notes",
		ContentMD:    "# Kernel Notes\n\nLLMS test.",
		Excerpt:      "Kernel notes for LLMS testing.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	_, err = state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "LLMS Test Project",
		ContentMD: "# LLMS Project\n\nProject for LLMS test.",
		Excerpt:   "LLMS test project description.",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	text, err := state.svc.SiteMeta.LLMSText()
	if err != nil {
		t.Fatalf("LLMSText: %v", err)
	}

	if text == "" {
		t.Fatalf("LLMSText should not be empty")
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "raevtar") {
		t.Fatalf("LLMSText = %q, want to contain 'raevtar'", text)
	}

	if !strings.Contains(lower, "llms kernel notes") {
		t.Fatalf("LLMSText = %q, want to contain post title 'llms kernel notes'", text)
	}

	if !strings.Contains(lower, "llms test project") {
		t.Fatalf("LLMSText = %q, want to contain project title 'llms test project'", text)
	}
}

func TestSiteMetaServiceLLMSTextNoContent(t *testing.T) {
	state := newTestServices(t)

	text, err := state.svc.SiteMeta.LLMSText()
	if err != nil {
		t.Fatalf("LLMSText: %v", err)
	}

	if text == "" {
		t.Fatalf("LLMSText should not be empty even without content")
	}

	lower := strings.ToLower(text)
	if !strings.Contains(lower, "raevtar") {
		t.Fatalf("LLMSText = %q, want to contain 'raevtar'", text)
	}

	if strings.Contains(lower, "recent blog posts") {
		t.Fatalf("LLMSText should not contain 'recent blog posts' section when no posts exist")
	}

	if strings.Contains(lower, "public projects") {
		t.Fatalf("LLMSText should not contain 'public projects' section when no projects exist")
	}
}

func TestSiteMetaServiceDefaultSEOPreservesLabDocsPath(t *testing.T) {
	state := newTestServices(t)

	seo := state.svc.SiteMeta.DefaultSEO("/lab/docs")

	if seo.CanonicalURL != "https://raevtar.test/docs" {
		t.Fatalf("DefaultSEO(/lab/docs) CanonicalURL = %q, want https://raevtar.test/docs (should normalize to /docs)", seo.CanonicalURL)
	}
}
