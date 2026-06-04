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
	Editorial *EditorialInboxService
	Media     *MediaService
	Monitor   *MonitorService
	Admin     *AdminService
}

func New(repos *repo.Repositories, cfg *config.Config) *Service {
	return &Service{
		repos:     repos,
		Cfg:       cfg,
		Blog:      NewBlogService(repos),
		Editorial: NewEditorialInboxService(repos),
		Media:     NewMediaService(repos, cfg.MediaDir),
		Monitor:   &MonitorService{repos: repos},
		Admin:     &AdminService{repos: repos},
	}
}
