package service

import (
	"log/slog"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

// SeedData populates initial categories and admin user.
func (s *Service) SeedData() error {
	// Seed categories
	categories := []model.Category{
		{Slug: "ai-agent", Name: "AI Agent", Description: "Proyek AI agent wajib coba"},
		{Slug: "security", Name: "Security", Description: "Tools keamanan & penetration testing"},
		{Slug: "kernel-embedded", Name: "Kernel & Embedded", Description: "Mainline kernel, postmarketOS, low-level"},
		{Slug: "devops", Name: "DevOps", Description: "Infrastructure, automation, CI/CD"},
		{Slug: "tools", Name: "Tools", Description: "Aplikasi & utilities keren"},
	}
	for _, cat := range categories {
		if err := s.Repos.Category.Create(&cat); err != nil {
			return err
		}
	}

	// Seed admin user from config if no users exist
	count, err := s.Repos.User.Count()
	if err != nil {
		return err
	}
	if count == 0 && s.Cfg.AdminPass != "" {
		hash, err := repo.HashPassword(s.Cfg.AdminPass)
		if err != nil {
			return err
		}
		u, err := s.Repos.User.Create(s.Cfg.AdminUser, hash, model.RoleOwner, s.Cfg.AdminUser)
		if err != nil {
			return err
		}
		slog.Info("seeded admin user", "username", u.Username, "role", u.Role)
	}

	return nil
}
