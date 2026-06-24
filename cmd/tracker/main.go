package main

import (
	"log"

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

	ui.GetAllDaysInMonth(6)
	// p := tea.NewProgram(ui.InitialModel())
	// if _, err = p.Run(); err != nil {
	// 	fmt.Printf("Alas, there error: %v", err)
	// 	os.Exit(1)
	// }
}
