# Containerized CLI Login System

Secure interactive CLI login system written in Go. It supports registration, password authentication, optional TOTP-based 2FA, account lockout, session timeout, and SQLite persistence through Docker volumes.

## Features

- User registration with unique usernames
- Password hashing with bcrypt
- Login with username and password
- Optional TOTP 2FA compatible with Google Authenticator
- Account lockout after repeated failed login attempts
- In-memory sessions with configurable expiration
- Interactive CLI with history and tab completion
- SQLite database persistence
- Docker and Docker Compose setup

## Commands

Before login:

```text
register
login
help
exit
```

After login:

```text
whoami
enable-2fa
disable-2fa
logout
help
exit
```

## Configuration

The application uses environment variables with these defaults:

```text
DB_PATH=./data/app.db
SESSION_TIMEOUT_MINUTES=30
MAX_FAILED_ATTEMPTS=5
LOCKOUT_MINUTES=15
```

In Docker, the database path is set to:

```text
/app/data/app.db
```

The `/app/data` directory is mounted as a Docker volume so data persists across container restarts.

## Run With Docker Compose

Build and start the CLI:

```bash
docker compose run --rm cli-login
```

If your Docker installation uses the older Compose command:

```bash
docker-compose run --rm cli-login
```

## Run Locally

Install Go 1.22 or newer, then run:

```bash
go mod tidy
go run ./cmd/cli
```

## Example Flow

```text
> register
Username: alice
Password:
User "alice" registered successfully.

> login
Username: alice
Password:
Login successful.
User details:
Username: alice
Registration date: ...
MFA status: disabled
Session expiration: ...
Last login: never

> enable-2fa
Secret: ...
Provisioning URL: otpauth://...
Enter current TOTP code to confirm: 123456
2FA enabled successfully.

> logout
Logged out.
```

## Database Schema

The migration is included in `migrations/001_create_users.sql`.

Main fields:

- `username`
- `password_hash`
- `totp_secret`
- `mfa_enabled`
- `failed_attempts`
- `locked_until`
- `created_at`
- `last_login_at`

## Tests

Run tests with:

```bash
go test ./...
```

## Security Notes

- Passwords are never stored in plain text.
- Failed password and failed TOTP attempts count toward lockout.
- TOTP secrets are only stored after the user confirms setup with a valid code.
- Sessions are in memory and expire automatically based on configuration.
