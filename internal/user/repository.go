package user

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(username, passwordHash string) (*User, error) {
	now := time.Now().UTC()
	result, err := r.db.Exec(`
		INSERT INTO users (username, password_hash, created_at)
		VALUES (?, ?, ?)
	`, strings.TrimSpace(username), passwordHash, now)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read created user id: %w", err)
	}

	return r.ByID(id)
}

func (r *Repository) ByUsername(username string) (*User, error) {
	return r.scanOne(`
		SELECT id, username, password_hash, COALESCE(totp_secret, ''), mfa_enabled,
		       failed_attempts, locked_until, created_at, last_login_at
		FROM users
		WHERE username = ?
	`, strings.TrimSpace(username))
}

func (r *Repository) ByID(id int64) (*User, error) {
	return r.scanOne(`
		SELECT id, username, password_hash, COALESCE(totp_secret, ''), mfa_enabled,
		       failed_attempts, locked_until, created_at, last_login_at
		FROM users
		WHERE id = ?
	`, id)
}

func (r *Repository) RecordFailedLogin(id int64, attempts int, lockedUntil *time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET failed_attempts = ?, locked_until = ?
		WHERE id = ?
	`, attempts, lockedUntil, id)
	if err != nil {
		return fmt.Errorf("record failed login: %w", err)
	}
	return nil
}

func (r *Repository) RecordSuccessfulLogin(id int64, lastLogin time.Time) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET failed_attempts = 0, locked_until = NULL, last_login_at = ?
		WHERE id = ?
	`, lastLogin, id)
	if err != nil {
		return fmt.Errorf("record successful login: %w", err)
	}
	return nil
}

func (r *Repository) EnableMFA(id int64, secret string) error {
	_, err := r.db.Exec(`UPDATE users SET totp_secret = ?, mfa_enabled = 1 WHERE id = ?`, secret, id)
	if err != nil {
		return fmt.Errorf("enable mfa: %w", err)
	}
	return nil
}

func (r *Repository) DisableMFA(id int64) error {
	_, err := r.db.Exec(`UPDATE users SET totp_secret = NULL, mfa_enabled = 0 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("disable mfa: %w", err)
	}
	return nil
}

func (r *Repository) scanOne(query string, args ...any) (*User, error) {
	var u User
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime

	err := r.db.QueryRow(query, args...).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.TOTPSecret,
		&u.MFAEnabled,
		&u.FailedAttempts,
		&lockedUntil,
		&u.CreatedAt,
		&lastLoginAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	if lockedUntil.Valid {
		u.LockedUntil = &lockedUntil.Time
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}

	return &u, nil
}
