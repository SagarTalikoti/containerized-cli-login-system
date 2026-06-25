package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cli-login-system/internal/user"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account is temporarily locked")
	ErrTOTPRequired       = errors.New("totp code required")
	ErrInvalidTOTP        = errors.New("invalid totp code")
	ErrUsernameRequired   = errors.New("username is required")
	ErrPasswordRequired   = errors.New("password is required")
)

type Service struct {
	users             *user.Repository
	maxFailedAttempts int
	lockoutDuration   time.Duration
}

func NewService(users *user.Repository, maxFailedAttempts int, lockoutDuration time.Duration) *Service {
	return &Service{users: users, maxFailedAttempts: maxFailedAttempts, lockoutDuration: lockoutDuration}
}

func (s *Service) Register(username, password string) (*user.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, ErrUsernameRequired
	}
	if password == "" {
		return nil, ErrPasswordRequired
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	return s.users.Create(username, hash)
}

func (s *Service) Login(username, password, totpCode string) (*user.User, error) {
	u, err := s.users.ByUsername(username)
	if errors.Is(err, user.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if u.LockedUntil != nil && now.Before(*u.LockedUntil) {
		return nil, fmt.Errorf("%w until %s", ErrAccountLocked, u.LockedUntil.Local().Format(time.RFC1123))
	}

	if !CheckPassword(u.PasswordHash, password) {
		if err := s.recordFailure(u, now); err != nil {
			return nil, err
		}
		return nil, ErrInvalidCredentials
	}

	if u.MFAEnabled {
		if strings.TrimSpace(totpCode) == "" {
			return nil, ErrTOTPRequired
		}
		if !ValidateTOTP(strings.TrimSpace(totpCode), u.TOTPSecret) {
			if err := s.recordFailure(u, now); err != nil {
				return nil, err
			}
			return nil, ErrInvalidTOTP
		}
	}

	previousLastLogin := u.LastLoginAt
	if err := s.users.RecordSuccessfulLogin(u.ID, now); err != nil {
		return nil, err
	}

	updated, err := s.users.ByID(u.ID)
	if err != nil {
		return nil, err
	}
	updated.LastLoginAt = previousLastLogin
	return updated, nil
}

func (s *Service) EnableMFA(u *user.User, code string, key *TOTPKey) error {
	if key == nil {
		return errors.New("totp key is required")
	}
	if !ValidateTOTP(strings.TrimSpace(code), key.Secret) {
		return ErrInvalidTOTP
	}
	return s.users.EnableMFA(u.ID, key.Secret)
}

func (s *Service) DisableMFA(u *user.User, code string) error {
	if u.MFAEnabled && !ValidateTOTP(strings.TrimSpace(code), u.TOTPSecret) {
		return ErrInvalidTOTP
	}
	return s.users.DisableMFA(u.ID)
}

func (s *Service) recordFailure(u *user.User, now time.Time) error {
	attempts := u.FailedAttempts + 1
	var lockedUntil *time.Time
	if attempts >= s.maxFailedAttempts {
		locked := now.Add(s.lockoutDuration)
		lockedUntil = &locked
	}
	return s.users.RecordFailedLogin(u.ID, attempts, lockedUntil)
}
