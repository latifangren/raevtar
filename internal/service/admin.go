package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type AdminService struct {
	repos *repo.Repositories
}

func (s *AdminService) ListUsers() ([]model.User, error) {
	users, err := s.repos.User.List()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *AdminService) Authenticate(username, password, ip string) (*model.User, error) {
	user, err := s.repos.User.GetByUsername(username)
	if err != nil {
		_ = s.repos.Audit.Insert(username, "LOGIN_FAILED", "user not found", ip)
		return nil, fmt.Errorf("authenticate user: %w", err)
	}
	if !repo.CheckPassword(password, user.PasswordHash) {
		_ = s.repos.Audit.Insert(username, "LOGIN_FAILED", "wrong password", ip)
		return nil, fmt.Errorf("authenticate user: invalid password")
	}
	if err := s.repos.Audit.Insert(user.Username, "LOGIN", "login via admin panel", ip); err != nil {
		return nil, fmt.Errorf("audit login: %w", err)
	}
	return user, nil
}

func (s *AdminService) LogLogout(username, ip string) error {
	if username == "" {
		return nil
	}
	if err := s.repos.Audit.Insert(username, "LOGOUT", "manual logout", ip); err != nil {
		return fmt.Errorf("audit logout: %w", err)
	}
	return nil
}

func (s *AdminService) CreateUser(actorRole, actorUsername, username, password, role, ip string) (*model.User, error) {
	if !model.IsValidRole(role) {
		role = model.RoleOperator
	}
	if !model.CanManage(actorRole, role) {
		return nil, fmt.Errorf("create user: forbidden role")
	}
	hash, err := repo.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repos.User.Create(username, hash, role, username)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if err := s.repos.Audit.Insert(actorUsername, "CREATE_USER", "created user "+user.Username+" with role "+user.Role, ip); err != nil {
		return nil, fmt.Errorf("audit create user: %w", err)
	}
	return user, nil
}

func (s *AdminService) DeleteUser(actorRole, actorUsername string, id int64, ip string) error {
	target, err := s.repos.User.GetByID(id)
	if err != nil {
		return fmt.Errorf("get target user: %w", err)
	}
	if !model.CanManage(actorRole, target.Role) {
		return fmt.Errorf("delete user: forbidden")
	}
	if err := s.repos.Audit.Insert(actorUsername, "DELETE_USER", "deleted user: "+target.Username+" (role: "+target.Role+")", ip); err != nil {
		return fmt.Errorf("audit delete user: %w", err)
	}
	if err := s.repos.User.Delete(id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *AdminService) ListAuditLogs(limit, offset int) ([]model.AuditLog, error) {
	logs, err := s.repos.Audit.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}

func (s *AdminService) ListServerAuditLogs(serverID int64, limit int) ([]model.AuditLog, error) {
	logs, err := s.repos.Audit.ListServerLogs(strconv.FormatInt(serverID, 10), limit)
	if err != nil {
		return nil, fmt.Errorf("list server audit logs: %w", err)
	}
	return logs, nil
}

func (s *AdminService) LogPostCreated(username, title, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_POST", "created post: "+title, ip); err != nil {
		return fmt.Errorf("audit create post: %w", err)
	}
	return nil
}

func (s *AdminService) LogPostUpdated(username, title, ip string) error {
	if err := s.repos.Audit.Insert(username, "UPDATE_POST", "updated post: "+title, ip); err != nil {
		return fmt.Errorf("audit update post: %w", err)
	}
	return nil
}

func (s *AdminService) DeletePost(username string, id int64, ip string) error {
	post, err := s.repos.Post.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrPostNotFound, err)
		}
		return fmt.Errorf("get post: %w", err)
	}
	if err := s.repos.Audit.Insert(username, "DELETE_POST", "deleted post: "+post.Title, ip); err != nil {
		return fmt.Errorf("audit delete post: %w", err)
	}
	if err := s.repos.Post.Delete(id); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func (s *AdminService) LogProjectCreated(username, title, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_PROJECT", "created project: "+title, ip); err != nil {
		return fmt.Errorf("audit create project: %w", err)
	}
	return nil
}

func (s *AdminService) LogProjectUpdated(username, title, ip string) error {
	if err := s.repos.Audit.Insert(username, "UPDATE_PROJECT", "updated project: "+title, ip); err != nil {
		return fmt.Errorf("audit update project: %w", err)
	}
	return nil
}

func (s *AdminService) LogCategoryCreated(username, name, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_CATEGORY", "created category: "+name, ip); err != nil {
		return fmt.Errorf("audit create category: %w", err)
	}
	return nil
}

func (s *AdminService) LogCategoryUpdated(username, name, ip string) error {
	if err := s.repos.Audit.Insert(username, "UPDATE_CATEGORY", "updated category: "+name, ip); err != nil {
		return fmt.Errorf("audit update category: %w", err)
	}
	return nil
}

func (s *AdminService) DeleteCategory(username string, category *model.Category, ip string) error {
	if err := s.repos.Audit.Insert(username, "DELETE_CATEGORY", "deleted category: "+category.Name, ip); err != nil {
		return fmt.Errorf("audit delete category: %w", err)
	}
	return nil
}

func (s *AdminService) DeleteProject(username string, id int64, ip string) error {
	project, err := s.repos.Project.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return fmt.Errorf("get project: %w", err)
	}
	if err := s.repos.Audit.Insert(username, "DELETE_PROJECT", "deleted project: "+project.Title, ip); err != nil {
		return fmt.Errorf("audit delete project: %w", err)
	}
	if err := s.repos.Project.Delete(id); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func (s *AdminService) LogPageUpdated(username string, page *model.PageContent, ip string) error {
	if err := s.repos.Audit.Insert(username, "UPDATE_PAGE_CONTENT", "updated page: "+page.Key, ip); err != nil {
		return fmt.Errorf("audit update page: %w", err)
	}
	return nil
}

func (s *AdminService) LogServerCreated(username, name, host, port, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_SERVER", "created server: "+name+" ("+host+":"+port+")", ip); err != nil {
		return fmt.Errorf("audit create server: %w", err)
	}
	return nil
}

func (s *AdminService) LogServerUpdated(username string, serverID int64, name, host, port, ip string) error {
	details := "updated server id: " + strconv.FormatInt(serverID, 10) + " (" + name + " " + host + ":" + port + ")"
	if err := s.repos.Audit.Insert(username, "UPDATE_SERVER", details, ip); err != nil {
		return fmt.Errorf("audit update server: %w", err)
	}
	return nil
}

func (s *AdminService) LogAgentTokenRotated(username, serverID, ip string) error {
	if err := s.repos.Audit.Insert(username, "ROTATE_AGENT_TOKEN", "rotated agent token for server id: "+serverID, ip); err != nil {
		return fmt.Errorf("audit rotate agent token: %w", err)
	}
	return nil
}

func (s *AdminService) DeleteServer(username string, id int64, idText, ip string) error {
	if _, err := s.repos.Server.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrServerNotFound, err)
		}
		return fmt.Errorf("get server: %w", err)
	}
	if err := s.repos.Audit.Insert(username, "DELETE_SERVER", "deleted server id: "+idText, ip); err != nil {
		return fmt.Errorf("audit delete server: %w", err)
	}
	if err := s.repos.Server.Delete(id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	return nil
}

func (s *AdminService) LogEditorialInboxCreated(username string, item *model.EditorialInboxItem, ip string) error {
	details := "created editorial item #" + strconv.FormatInt(item.ID, 10) + " (" + item.SourceType + ": " + item.SourceValue + ")"
	if err := s.repos.Audit.Insert(username, "CREATE_EDITORIAL_ITEM", details, ip); err != nil {
		return fmt.Errorf("audit create editorial item: %w", err)
	}
	return nil
}

func (s *AdminService) LogEditorialInboxUpdated(username string, item *model.EditorialInboxItem, ip string) error {
	details := "updated editorial item #" + strconv.FormatInt(item.ID, 10) + " status=" + item.Status + " mode=" + item.Mode
	if err := s.repos.Audit.Insert(username, "UPDATE_EDITORIAL_ITEM", details, ip); err != nil {
		return fmt.Errorf("audit update editorial item: %w", err)
	}
	return nil
}

func (s *AdminService) LogEditorialInboxDeleted(username string, item *model.EditorialInboxItem, ip string) error {
	details := "deleted editorial item #" + strconv.FormatInt(item.ID, 10) + " (" + item.SourceType + ": " + item.SourceValue + ")"
	if err := s.repos.Audit.Insert(username, "DELETE_EDITORIAL_ITEM", details, ip); err != nil {
		return fmt.Errorf("audit delete editorial item: %w", err)
	}
	return nil
}

func (s *AdminService) LogAudit(username, action, details, ip string) error {
	if err := s.repos.Audit.Insert(username, action, details, ip); err != nil {
		return fmt.Errorf("audit log: %w", err)
	}
	return nil
}
