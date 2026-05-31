package service

import (
	"fmt"

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

func (s *AdminService) LogPostCreated(username, title, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_POST", "created post: "+title, ip); err != nil {
		return fmt.Errorf("audit create post: %w", err)
	}
	return nil
}

func (s *AdminService) DeletePost(username string, id int64, ip string) error {
	post, err := s.repos.Post.GetByID(id)
	if err == nil {
		if err := s.repos.Audit.Insert(username, "DELETE_POST", "deleted post: "+post.Title, ip); err != nil {
			return fmt.Errorf("audit delete post: %w", err)
		}
	}
	if err := s.repos.Post.Delete(id); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func (s *AdminService) LogServerCreated(username, name, host, port, ip string) error {
	if err := s.repos.Audit.Insert(username, "CREATE_SERVER", "created server: "+name+" ("+host+":"+port+")", ip); err != nil {
		return fmt.Errorf("audit create server: %w", err)
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
	if err := s.repos.Audit.Insert(username, "DELETE_SERVER", "deleted server id: "+idText, ip); err != nil {
		return fmt.Errorf("audit delete server: %w", err)
	}
	if err := s.repos.Server.Delete(id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	return nil
}
