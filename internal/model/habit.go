// Package model - test comment for linter
package model

import "time"

type HabitType string

const (
	HabitProgress HabitType = "progress"
	HabitCount    HabitType = "count"
	HabitMinutes  HabitType = "minutes"
)

type Habit struct {
	ID        int64
	Name      string
	Type      HabitType
	CreatedAt time.Time
	Goal      *float64
	StartDate time.Time
	EndDate   *time.Time
}

type Entry struct {
	ID      int64
	HabitID int64
	Date    time.Time
	Value   float64
}
