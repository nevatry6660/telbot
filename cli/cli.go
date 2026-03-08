package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
	"telkomsel-bot/telkomsel"
)

func Run() {

	logPath := filepath.Join(os.TempDir(), "telkomsel-cli.log")
	logFile, err := os.Create(logPath)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		log.SetOutput(io.Discard)
	}

	api := telkomsel.NewClient()
	defer api.Close()
	auth := telkomsel.NewAuth()
	sessions := model.NewSessionManager(config.GetSessionPath())

	var loggedInUser *model.Session
	var loggedInID int64
	for id, s := range sessions.All() {
		if s.IsLoggedIn() {
			if loggedInUser == nil || s.LastLoginAt.After(loggedInUser.LastLoginAt) {
				loggedInUser = s
				loggedInID = id
			}
		}
	}

	m := newModel(api, auth, sessions)
	m.loggedInUser = loggedInUser
	m.loggedInID = loggedInID

	p := tea.NewProgram(m, tea.WithAltScreen())
	programRef = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
