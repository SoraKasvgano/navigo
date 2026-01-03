package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VerifyPassword 验证密码
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	err := db.QueryRow(
		"SELECT id, username, password, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateDefaultUser 创建默认管理员用户
func CreateDefaultUser(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hashedPassword, err := HashPassword("admin")
		if err != nil {
			return err
		}

		_, err = db.Exec(
			"INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)",
			"admin", hashedPassword, time.Now(), time.Now(),
		)
		return err
	}
	return nil
}

// UpdatePassword 更新用户密码
func UpdatePassword(db *sql.DB, username string, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	result, err := db.Exec(
		"UPDATE users SET password = ?, updated_at = ? WHERE username = ?",
		hashedPassword, time.Now(), username,
	)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected != 1 {
		return sql.ErrNoRows
	}
	return nil
}
