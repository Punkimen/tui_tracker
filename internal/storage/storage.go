// Package storage for work with sqlite
package storage

import (
	"time"

	"daily-tracker/internal/model"
)

type Storage interface {
	SaveHabit(h model.Habit) (int64, error)
	GetHabits() ([]model.Habit, error)
	DeleteHabit(id int64) error
	HabitEndDate(id int64, date time.Time) error

	SaveEntry(e model.Entry) error
	GetEntries() ([]model.Entry, error)
	GetEntriesByMonth(year int, month time.Month) ([]model.Entry, error)
	UpdateEntry(id int64, value float64) (model.Entry, error)
	UpsertEntry(habitID int64, date time.Time, value float64) error
	DeleteEntry(id int64) error
}
