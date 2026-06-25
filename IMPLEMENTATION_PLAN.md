# Go Backend Task Implementation Plan

## Objective

Build a secure command-line login system in Go that supports user registration, authentication, optional TOTP-based 2FA, session management, Docker containerization, and persistent database storage.

## Recommended Stack

- Language: Go
- Database: SQLite
- Password hashing: bcrypt
- 2FA: TOTP, Google Authenticator compatible
- CLI: interactive prompt with history and tab completion
- Containerization: Docker and Docker Compose
- Persistence: SQLite database stored in a Docker volume

## Proposed Project Structure

```text
.
├── cmd/
│   └── cli/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── service.go
│   │   ├── password.go
│   │   ├── totp.go
│   │   └── session.go
│   ├── cli/
│   │   ├── prompt.go
│   │   └── commands.go
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── sqlite.go
│   │   └── migrations.go
│   └── user/
│       ├── model.go
│       └── repository.go
├── migrations/
│   └── 001_create_users.sql
├── Dockerfile
├── docker-compose.yml
├── README.md
├── go.mod
└── go.sum
```

## Database Design

Use SQLite for simplicity and persist the database file through a Docker volume.

### Users Table

```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  totp_secret TEXT,
  mfa_enabled BOOLEAN NOT NULL DEFAULT 0,
  failed_attempts INTEGER NOT NULL DEFAULT 0,
  locked_until DATETIME,
  created_at DATETIME NOT NULL,
  last_login_at DATETIME
);
```

Session data can remain in memory because this is an interactive CLI application.

## Required Commands

### Before Login

- `register` - create a new user
- `login` - authenticate with username and password, plus TOTP if enabled
- `help` - show available commands
- `exit` - quit the program

### After Login

- `whoami` - show current user details
- `enable-2fa` - enable TOTP-based MFA
- `disable-2fa` - disable MFA
- `logout` - end current session
- `help` - show available commands
- `exit` - quit the program

## Feature Plan

### Registration

- Prompt for username.
- Prompt for password securely without echoing input.
- Validate username uniqueness.
- Hash password with bcrypt.
- Store user in SQLite.
- Display clear success or error feedback.

### Login

- Prompt for username and password.
- Check account lockout before password verification.
- Verify password using bcrypt.
- If MFA is enabled, prompt for TOTP code.
- On successful login:
  - Reset failed attempts.
  - Clear lockout value.
  - Update last login time.
  - Start an in-memory session.
  - Display user details automatically.

### Account Lockout

- Track failed login attempts per user.
- Lock account after a configurable number of failed attempts.
- Recommended defaults:
  - Maximum failed attempts: `5`
  - Lockout duration: `15 minutes`
- Store `failed_attempts` and `locked_until` in the database.

### Session Management

- Keep current session in memory.
- Store username, user ID, login time, and expiration time.
- Use configurable session timeout.
- Check session validity before every authenticated command.
- Automatically logout and notify the user when the session expires.

### TOTP 2FA

- `enable-2fa`:
  - Generate a TOTP secret.
  - Show the secret and provisioning URL for Google Authenticator.
  - Ask the user to enter a current TOTP code to confirm setup.
  - Enable MFA only after successful verification.
- `disable-2fa`:
  - Require verification, preferably current TOTP code or password.
  - Disable MFA and clear the stored TOTP secret.

### User Details After Login

Display the following immediately after successful login:

- Username
- Registration date
- MFA status, enabled or disabled
- Session expiration time
- Last login time, if available

## Configuration

Use environment variables with sensible defaults.

```text
DB_PATH=/app/data/app.db
SESSION_TIMEOUT_MINUTES=30
MAX_FAILED_ATTEMPTS=5
LOCKOUT_MINUTES=15
```

## Docker Plan

### Dockerfile

- Use a Go build stage to compile the CLI binary.
- Use a minimal runtime image.
- Copy the compiled binary into the runtime image.
- Set the default command to run the CLI.

### docker-compose.yml

- Build and run the CLI application container.
- Mount a Docker volume for SQLite persistence.
- Pass configuration through environment variables.

Recommended database path inside the container:

```text
/app/data/app.db
```

Recommended volume:

```yaml
volumes:
  cli_data:
```

## README Requirements

The README should include:

- Project overview
- Feature list
- Prerequisites
- Docker setup instructions
- How to run the CLI
- Full command reference
- Example usage flow
- Environment variable configuration
- Database persistence explanation
- Security notes
- Test instructions, if tests are included

## Optional Unit Tests

Add tests for core behavior if time allows:

- Password hashing and verification
- TOTP generation and validation
- Session expiration
- Account lockout rules
- User repository operations

## Implementation Order

1. Initialize Go module.
2. Add required dependencies.
3. Create database connection and migration logic.
4. Implement user model and repository.
5. Implement password hashing and verification.
6. Implement registration flow.
7. Implement login flow.
8. Implement failed-attempt tracking and lockout.
9. Implement session management.
10. Implement TOTP enable and disable flow.
11. Build interactive CLI with command history and tab completion.
12. Add Dockerfile and docker-compose.yml.
13. Write README.md.
14. Add optional unit tests.
15. Verify the full flow locally and through Docker Compose.

## Final Verification Checklist

- User can register successfully.
- User cannot register with duplicate username.
- User can login with valid credentials.
- Invalid login attempts increment failure count.
- Account lockout works after repeated failed attempts.
- Session expires after configured timeout.
- `whoami` displays required user details.
- 2FA can be enabled and verified.
- Login requires TOTP after 2FA is enabled.
- 2FA can be disabled.
- Logout clears the session.
- CLI supports help, history, and tab completion.
- Docker Compose runs the application.
- SQLite data persists across container restarts.
- README contains setup and usage documentation.
