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
		if err := s.repos.Category.Create(&cat); err != nil {
			return err
		}
	}

	// Seed admin user from config if no users exist
	count, err := s.repos.User.Count()
	if err != nil {
		return err
	}
	if count == 0 && s.Cfg.AdminPass != "" {
		hash, err := repo.HashPassword(s.Cfg.AdminPass)
		if err != nil {
			return err
		}
		u, err := s.repos.User.Create(s.Cfg.AdminUser, hash, model.RoleOwner, s.Cfg.AdminUser)
		if err != nil {
			return err
		}
		slog.Info("seeded admin user", "username", u.Username, "role", u.Role)
	}

	defaults := []model.PageContent{
		{
			Key:       model.PageKeyAbout,
			Title:     "Single binary, public-safe by design.",
			Summary:   "Raevtar keeps writing, monitoring, and automation inside one small Go binary while drawing a hard line between public signals and operator-only detail.",
			ContentMD: "Raevtar is personal platform for writing notes, watching machines, and wiring light automation together in one Go binary. Public pages stay useful, but operator-only details stay backstage.",
		},
		{
			Key:       model.PageKeyContact,
			Title:     "Raevtar keeps contact intentionally lightweight.",
			Summary:   "Best first contact here is context, not a form. Start with public writing, route docs, and status board before moving elsewhere.",
			ContentMD: "Best first contact here is context, not a form. Start with public writing, route docs, and status board, then carry shared context wherever conversation continues off-site.",
		},
	}
	for _, page := range defaults {
		if err := s.repos.PageContent.Upsert(&page); err != nil {
			return err
		}
	}

	return nil
}
