package main

import (
	"fmt"
	"os"

	"cli-login-system/internal/auth"
	"cli-login-system/internal/cli"
	"cli-login-system/internal/config"
	"cli-login-system/internal/db"
	"cli-login-system/internal/user"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database error:", err)
		os.Exit(1)
	}
	defer database.Close()

	users := user.NewRepository(database)
	authService := auth.NewService(users, cfg.MaxFailedAttempts, cfg.LockoutDuration)
	sessions := auth.NewSessionManager(cfg.SessionTimeout)

	app, err := cli.NewApp(authService, sessions, users)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cli error:", err)
		os.Exit(1)
	}

	app.Run()
}
