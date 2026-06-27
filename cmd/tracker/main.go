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

	ui.GetDaysFromMoth(5)
}
