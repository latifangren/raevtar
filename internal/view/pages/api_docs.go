package pages

const (
	postsListJSON = `{
  "data": [
    {
      "id": 1,
      "title": "Judul Artikel",
      "slug": "judul-artikel",
      "excerpt": "Ringkasan artikel",
      "category_id": 1,
      "category_slug": "ai-agent",
      "category_name": "AI Agent",
      "cover_image_url": "/static/uploads/cover.jpg",
      "tags": [
        {"id": 1, "name": "AI", "slug": "ai"}
      ],
      "published": true,
      "created_at": "2026-06-23T10:00:00Z",
      "updated_at": "2026-06-23T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}`

	postDetailJSON = `{
  "id": 1,
  "title": "Judul Artikel",
  "slug": "judul-artikel",
  "excerpt": "Ringkasan artikel",
  "content_md": "## Markdown content\n\nParagraph here...",
  "content_html": "<h2>Markdown content</h2>\n<p>Paragraph here...</p>",
  "category_id": 1,
  "category_slug": "ai-agent",
  "category_name": "AI Agent",
  "cover_image_url": "/static/uploads/cover.jpg",
  "tags": [
    {"id": 1, "name": "AI", "slug": "ai"}
  ],
  "published": true,
  "created_at": "2026-06-23T10:00:00Z",
  "updated_at": "2026-06-23T10:00:00Z"
}`

	projectsListJSON = `{
  "data": [
    {
      "id": 1,
      "name": "Project Name",
      "slug": "project-name",
      "description": "Project description",
      "url": "https://github.com/raevtar/project",
      "tags": ["go", "htmx"],
      "featured": true,
      "sort_order": 1,
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-23T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20
}`

	categoriesListJSON = `[
  {
    "id": 1,
    "name": "AI Agent",
    "slug": "ai-agent",
    "description": "Posts about AI agents"
  }
]`
)
