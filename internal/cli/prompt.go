package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"golang.org/x/term"

	"cli-login-system/internal/auth"
	"cli-login-system/internal/user"
)

type App struct {
	auth     *auth.Service
	sessions *auth.SessionManager
	users    *user.Repository
	rl       *readline.Instance
}

func NewApp(authService *auth.Service, sessions *auth.SessionManager, users *user.Repository) (*App, error) {
	app := &App{auth: authService, sessions: sessions, users: users}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     historyFile(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    app.completer(),
	})
	if err != nil {
		return nil, err
	}

	app.rl = rl
	return app, nil
}

func historyFile() string {
	return os.TempDir() + string(os.PathSeparator) + "cli-login-history.tmp"
}

func (a *App) Run() {
	defer a.rl.Close()

	fmt.Println("Secure CLI Login System")
	fmt.Println("Type 'help' to list commands.")

	for {
		line, err := a.rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		}
		if err != nil {
			fmt.Println()
			return
		}

		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		if a.handle(command) {
			return
		}
	}
}

func (a *App) prompt(label string) (string, error) {
	a.rl.SetPrompt(label + ": ")
	defer a.rl.SetPrompt("> ")
	return a.rl.Readline()
}

func (a *App) promptPassword(label string) (string, error) {
	fmt.Print(label + ": ")
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *App) completer() readline.AutoCompleter {
	return readline.NewPrefixCompleter(
		readline.PcItem("register"),
		readline.PcItem("login"),
		readline.PcItem("whoami"),
		readline.PcItem("enable-2fa"),
		readline.PcItem("disable-2fa"),
		readline.PcItem("logout"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)
}

func formatTime(value *time.Time) string {
	if value == nil {
		return "never"
	}
	return value.Local().Format(time.RFC1123)
}
