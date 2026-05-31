package model

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role"` // owner | admin | operator | readonly
	DisplayName  string    `db:"display_name" json:"display_name"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type AuditLog struct {
	ID        int64     `db:"id" json:"id"`
	User      string    `db:"user" json:"user"`
	Action    string    `db:"action" json:"action"` // login, logout, create_post, delete_post, create_user, delete_user, etc
	Details   string    `db:"details" json:"details"`
	IP        string    `db:"ip" json:"ip"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Role constants
const (
	RoleOwner    = "owner"
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleReadonly = "readonly"
)

// Role hierarchy (lower index = more privilege)
var RoleLevels = map[string]int{
	RoleOwner:    0,
	RoleAdmin:    1,
	RoleOperator: 2,
	RoleReadonly: 3,
}

func RoleLevel(role string) int {
	if l, ok := RoleLevels[role]; ok {
		return l
	}
	return 999
}

// CanManage returns true if the given role can manage another user with targetRole.
// A user can only manage users of lower or equal privilege level.
func CanManage(role, targetRole string) bool {
	return RoleLevel(role) < RoleLevel(targetRole)
}

// ValidRoles returns all valid role strings.
func ValidRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleOperator, RoleReadonly}
}

// IsValidRole checks if a role string is valid.
func IsValidRole(role string) bool {
	_, ok := RoleLevels[role]
	return ok
}
