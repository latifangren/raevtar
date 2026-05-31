package service

import (
	"raevtar/internal/config"
	"raevtar/internal/repo"
)

// Service bundles all business logic.
// Handler calls Service. Service calls Repo.
type Service struct {
	Repos   *repo.Repositories
	Cfg     *config.Config
	Blog    *BlogService
	Monitor *MonitorService
}

func New(repos *repo.Repositories, cfg *config.Config) *Service {
	return &Service{
		Repos:   repos,
		Cfg:     cfg,
		Blog:    NewBlogService(repos),
		Monitor: &MonitorService{repos: repos},
	}
}
