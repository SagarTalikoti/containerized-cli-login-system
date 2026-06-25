package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cli-login-system/internal/auth"
	"cli-login-system/internal/user"
)

func (a *App) handle(command string) bool {
	switch strings.ToLower(command) {
	case "register":
		a.register()
	case "login":
		a.login()
	case "whoami":
		a.whoami()
	case "enable-2fa":
		a.enable2FA()
	case "disable-2fa":
		a.disable2FA()
	case "logout":
		a.logout()
	case "help":
		a.help()
	case "exit", "quit":
		return true
	default:
		fmt.Println("Unknown command. Type 'help' to list available commands.")
	}

	return false
}

func (a *App) register() {
	if _, ok := a.sessions.Current(); ok {
		fmt.Println("Logout before registering another user.")
		return
	}

	username, err := a.prompt("Username")
	if err != nil {
		fmt.Println("Could not read username:", err)
		return
	}
	password, err := a.promptPassword("Password")
	if err != nil {
		fmt.Println("Could not read password:", err)
		return
	}

	created, err := a.auth.Register(username, password)
	if err != nil {
		fmt.Println("Registration failed:", cleanError(err))
		return
	}

	fmt.Printf("User %q registered successfully.\n", created.Username)
}

func (a *App) login() {
	if _, ok := a.sessions.Current(); ok {
		fmt.Println("Already logged in. Use 'logout' first.")
		return
	}

	username, err := a.prompt("Username")
	if err != nil {
		fmt.Println("Could not read username:", err)
		return
	}
	password, err := a.promptPassword("Password")
	if err != nil {
		fmt.Println("Could not read password:", err)
		return
	}

	loggedIn, err := a.auth.Login(username, password, "")
	if errors.Is(err, auth.ErrTOTPRequired) {
		code, promptErr := a.prompt("TOTP code")
		if promptErr != nil {
			fmt.Println("Could not read TOTP code:", promptErr)
			return
		}
		loggedIn, err = a.auth.Login(username, password, code)
	}
	if err != nil {
		fmt.Println("Login failed:", cleanError(err))
		return
	}

	session := a.sessions.Start(loggedIn)
	fmt.Println("Login successful.")
	a.printUserDetails(loggedIn, session.ExpiresAt)
}

func (a *App) whoami() {
	session, ok := a.requireSession()
	if !ok {
		return
	}

	current, err := a.users.ByID(session.UserID)
	if err != nil {
		fmt.Println("Could not load user:", err)
		return
	}

	a.printUserDetails(current, session.ExpiresAt)
}

func (a *App) enable2FA() {
	session, ok := a.requireSession()
	if !ok {
		return
	}

	current, err := a.users.ByID(session.UserID)
	if err != nil {
		fmt.Println("Could not load user:", err)
		return
	}
	if current.MFAEnabled {
		fmt.Println("2FA is already enabled.")
		return
	}

	key, err := auth.GenerateTOTP(current.Username)
	if err != nil {
		fmt.Println("Could not generate TOTP secret:", err)
		return
	}

	fmt.Println("Add this account to Google Authenticator or another TOTP app.")
	fmt.Println("Secret:", key.Secret)
	fmt.Println("Provisioning URL:", key.URL)

	code, err := a.prompt("Enter current TOTP code to confirm")
	if err != nil {
		fmt.Println("Could not read TOTP code:", err)
		return
	}

	if err := a.auth.EnableMFA(current, code, key); err != nil {
		fmt.Println("Could not enable 2FA:", cleanError(err))
		return
	}

	fmt.Println("2FA enabled successfully.")
}

func (a *App) disable2FA() {
	session, ok := a.requireSession()
	if !ok {
		return
	}

	current, err := a.users.ByID(session.UserID)
	if err != nil {
		fmt.Println("Could not load user:", err)
		return
	}
	if !current.MFAEnabled {
		fmt.Println("2FA is already disabled.")
		return
	}

	code, err := a.prompt("Current TOTP code")
	if err != nil {
		fmt.Println("Could not read TOTP code:", err)
		return
	}

	if err := a.auth.DisableMFA(current, code); err != nil {
		fmt.Println("Could not disable 2FA:", cleanError(err))
		return
	}

	fmt.Println("2FA disabled successfully.")
}

func (a *App) logout() {
	if _, ok := a.sessions.Current(); !ok {
		fmt.Println("Not logged in.")
		return
	}
	a.sessions.Logout()
	fmt.Println("Logged out.")
}

func (a *App) help() {
	if _, ok := a.sessions.Current(); ok {
		fmt.Println("Available commands: whoami, enable-2fa, disable-2fa, logout, help, exit")
		return
	}
	fmt.Println("Available commands: register, login, help, exit")
}

func (a *App) requireSession() (*auth.Session, bool) {
	session, ok := a.sessions.Current()
	if !ok {
		fmt.Println("Not logged in or session expired. Use 'login' to continue.")
		return nil, false
	}
	return session, true
}

func (a *App) printUserDetails(u *user.User, expiresAt time.Time) {
	mfaStatus := "disabled"
	if u.MFAEnabled {
		mfaStatus = "enabled"
	}

	fmt.Println("User details:")
	fmt.Println("Username:", u.Username)
	fmt.Println("Registration date:", u.CreatedAt.Local().Format(time.RFC1123))
	fmt.Println("MFA status:", mfaStatus)
	fmt.Println("Session expiration:", expiresAt.Local().Format(time.RFC1123))
	fmt.Println("Last login:", formatTime(u.LastLoginAt))
}

func cleanError(err error) string {
	message := err.Error()
	if strings.Contains(message, "UNIQUE constraint failed") {
		return "username already exists"
	}
	return message
}
