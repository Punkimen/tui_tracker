package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"daily-tracker/internal/storage"
	"daily-tracker/internal/tracker"
	"daily-tracker/internal/ui"
)

func dataBasePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir = filepath.Join(dir, "daily-tracker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, "tracker.db"), nil
}

func main() {
	path, err := dataBasePath()
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.New(path)
	if err != nil {
		log.Fatal(err)
	}

	t := tracker.New(db)
	app := ui.CreateApp(t)

	p := tea.NewProgram(app)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
