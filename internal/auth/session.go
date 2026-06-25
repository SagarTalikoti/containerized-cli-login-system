package auth

import (
	"time"

	"cli-login-system/internal/user"
)

type Session struct {
	UserID    int64
	Username  string
	ExpiresAt time.Time
}

type SessionManager struct {
	current *Session
	timeout time.Duration
}

func NewSessionManager(timeout time.Duration) *SessionManager {
	return &SessionManager{timeout: timeout}
}

func (m *SessionManager) Start(u *user.User) *Session {
	// Sessions are intentionally in memory and expire after the configured timeout.
	m.current = &Session{
		UserID:    u.ID,
		Username:  u.Username,
		ExpiresAt: time.Now().UTC().Add(m.timeout),
	}
	return m.current
}

func (m *SessionManager) Current() (*Session, bool) {
	if m.current == nil {
		return nil, false
	}
	// Expired sessions are cleared before protected commands are allowed to run.
	if time.Now().UTC().After(m.current.ExpiresAt) {
		m.current = nil
		return nil, false
	}
	return m.current, true
}

func (m *SessionManager) Logout() {
	m.current = nil
}
