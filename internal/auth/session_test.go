package auth

import (
	"testing"
	"time"

	"cli-login-system/internal/user"
)

func TestSessionExpires(t *testing.T) {
	manager := NewSessionManager(time.Nanosecond)
	manager.Start(&user.User{ID: 1, Username: "alice"})

	time.Sleep(time.Millisecond)

	if _, ok := manager.Current(); ok {
		t.Fatal("expected session to expire")
	}
}
