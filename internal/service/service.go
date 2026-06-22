package service

import (
	"raevtar/internal/config"
	"raevtar/internal/repo"
)

// Service bundles all business logic.
// Handler calls Service. Service calls Repo.
type Service struct {
	repos     *repo.Repositories
	Cfg       *config.Config
	Blog      *BlogService
	Projects  *ProjectService
	Search    *SearchService
	SiteMeta  *SiteMetaService
	Pages     *PageContentService
	Editorial *EditorialInboxService
	Media     *MediaService
	Monitor   *MonitorService
	Admin     *AdminService
	CommandQ  *CommandQueueService
	Webhook   *WebhookService
}

func New(repos *repo.Repositories, cfg *config.Config) *Service {
	blog := NewBlogService(repos)
	projects := NewProjectService(repos)
	pages := NewPageContentService(repos)
	s := &Service{
		repos:     repos,
		Cfg:       cfg,
		Blog:      blog,
		Projects:  projects,
		Search:    NewSearchService(repos, blog, projects, pages),
		SiteMeta:  NewSiteMetaService(blog, projects, cfg.Domain),
		Pages:     pages,
		Editorial: NewEditorialInboxService(repos),
		Media:     NewMediaService(repos, cfg.MediaDir),
		Monitor:   &MonitorService{repos: repos},
		Admin:     &AdminService{repos: repos},
		CommandQ:  NewCommandQueueService(repos),
		Webhook:   NewWebhookService(repos),
	}
	s.Monitor.SetWebhook(s.Webhook)
	return s
}
