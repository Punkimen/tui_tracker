package main

import (
	"fmt"
	"log"
	"os"

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
	_ = t

	p := tea.NewProgram(ui.InitialModel())
	if _, err = p.Run(); err != nil {
		fmt.Printf("Alas, there error: %v", err)
		os.Exit(1)
	}
}
