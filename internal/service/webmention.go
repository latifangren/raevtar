package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type WebmentionService struct {
	repos *repo.Repositories
}

func NewWebmentionService(repos *repo.Repositories) *WebmentionService {
	return &WebmentionService{repos: repos}
}

type WebmentionInput struct {
	Source string
	Target string
}

func (s *WebmentionService) Receive(input WebmentionInput) (*model.Webmention, error) {
	source := strings.TrimSpace(input.Source)
	target := strings.TrimSpace(input.Target)
	if source == "" || target == "" {
		return nil, fmt.Errorf("webmention: source and target required")
	}

	// Extract post slug from target URL (e.g. https://raevtar.tech/blog/my-post)
	postID, err := s.resolvePostID(target)
	if err != nil {
		return nil, fmt.Errorf("webmention: %w", err)
	}

	// Fetch page info from source for enrichment
	title, author := s.fetchSourceInfo(source)

	w := &model.Webmention{
		SourceURL: source,
		TargetURL: target,
		PostID:    postID,
		Title:     title,
		Author:    author,
		Approved:  false, // require admin approval
	}
	if err := s.repos.Webmention.Create(w); err != nil {
		return nil, fmt.Errorf("webmention: save: %w", err)
	}
	slog.Info("webmention received", "source", source, "post_id", postID, "title", title)
	return w, nil
}

func (s *WebmentionService) ListByPost(postID int64, approvedOnly bool) ([]model.Webmention, error) {
	return s.repos.Webmention.ListByPost(postID, approvedOnly)
}

func (s *WebmentionService) ListAll(limit int) ([]model.Webmention, error) {
	return s.repos.Webmention.ListAll(limit)
}

func (s *WebmentionService) Approve(id int64) error {
	return s.repos.Webmention.Approve(id)
}

func (s *WebmentionService) Delete(id int64) error {
	return s.repos.Webmention.Delete(id)
}

func (s *WebmentionService) resolvePostID(targetURL string) (int64, error) {
	// Expect: https://domain/blog/{slug}
	parts := strings.Split(strings.TrimSuffix(targetURL, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid target URL: %s", targetURL)
	}
	slug := parts[len(parts)-1]
	if slug == "" {
		return 0, fmt.Errorf("empty slug in target URL: %s", targetURL)
	}
	post, err := s.repos.Post.GetBySlug(slug)
	if err != nil {
		return 0, fmt.Errorf("post not found for slug %q: %w", slug, err)
	}
	return post.ID, nil
}

func (s *WebmentionService) fetchSourceInfo(sourceURL string) (title, author string) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	// Read first 64KB for <title>
	buf := make([]byte, 65536)
	n, _ := resp.Body.Read(buf)
	html := string(buf[:n])

	// Extract <title>
	if idx := strings.Index(html, "<title"); idx >= 0 {
		start := strings.Index(html[idx:], ">")
		if start >= 0 {
			end := strings.Index(html[idx+start:], "</title>")
			if end >= 0 {
				title = strings.TrimSpace(html[idx+start+1 : idx+start+end])
			}
		}
	}

	return title, author
}
