package repo

import (
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo struct{ db *sqlx.DB }

func (r *UserRepo) List() ([]model.User, error) {
	var users []model.User
	return users, r.db.Select(&users, "SELECT * FROM users ORDER BY created_at ASC")
}

func (r *UserRepo) GetByID(id int64) (*model.User, error) {
	var u model.User
	err := r.db.Get(&u, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Get(&u, "SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(username, passwordHash, role, displayName string) (*model.User, error) {
	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO users (username, password_hash, role, display_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		username, passwordHash, role, displayName, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetByID(id)
}

func (r *UserRepo) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", passwordHash, time.Now(), id)
	return err
}

func (r *UserRepo) UpdateRole(id int64, role string) error {
	_, err := r.db.Exec("UPDATE users SET role = ?, updated_at = ? WHERE id = ?", role, time.Now(), id)
	return err
}

func (r *UserRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (r *UserRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM users")
	return count, err
}

// --- Password helpers ---

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
