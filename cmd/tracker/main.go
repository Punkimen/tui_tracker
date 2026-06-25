package main

import (
	"fmt"
	"os"

	"log"

	tea "charm.land/bubbletea/v2"
	"daily-tracker/internal/storage"
	"daily-tracker/internal/tracker"
	"daily-tracker/internal/ui"
)

func main() {
	db, err := storage.New("tracker.db")
	if err != nil {
		log.Fatal(err)
	}

	t := tracker.New(db)
	app := ui.NewApp(t)

	p := tea.NewProgram(app)
	if _, err = p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
