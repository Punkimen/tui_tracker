// Package tracker implements business logic for the habit tracker.
package tracker

import (
	"errors"
	"strings"
	"time"

	"daily-tracker/internal/model"
	"daily-tracker/internal/storage"
)

// Tracker — основная структура бизнес-логики.
// Хранит ссылку на storage через интерфейс, не через конкретный тип.
type Tracker struct {
	storage storage.Storage
}

type habitUpdater interface {
	UpdateHabit(h model.Habit) (model.Habit, error)
}

// New — конструктор Tracker. Принимает любую реализацию Storage.
// Аналог в JS: function createTracker(storage: Storage) { return { storage } }
func New(s storage.Storage) *Tracker {
	return &Tracker{storage: s}
}

// AddHabit создаёт новую привычку.
// time.Now() проставляем здесь — это бизнес-логика, не дело storage.
func (t *Tracker) AddHabit(
	name string,
	habitType model.HabitType,
	goal *float64,
) (model.Habit, error) {
	h := model.Habit{
		Name:      name,
		Type:      habitType,
		Goal:      goal,
		StartDate: time.Now(),
		CreatedAt: time.Now(),
	}

	id, err := t.storage.SaveHabit(h)
	if err != nil {
		return model.Habit{}, err
	}

	h.ID = id
	return h, nil
}

// GetHabits возвращает все привычки из хранилища.
func (t *Tracker) GetHabits(date time.Time) ([]model.Habit, error) {
	return t.storage.GetHabits(date)
}

func (t *Tracker) UpdateHabit(
	id int64,
	name string,
	habitType model.HabitType,
) (model.Habit, error) {
	if id <= 0 {
		return model.Habit{}, errors.New("habit id is invalid")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return model.Habit{}, errors.New("habit name is empty")
	}

	if habitType != model.HabitProgress &&
		habitType != model.HabitCount &&
		habitType != model.HabitMinutes {
		return model.Habit{}, errors.New("habit type is invalid")
	}

	updater, ok := t.storage.(habitUpdater)
	if !ok {
		return model.Habit{}, errors.New("storage does not support habit update")
	}

	habit, err := t.storage.GetHabitByID(id)
	if err != nil {
		return model.Habit{}, err
	}
	habit.Name = name
	habit.Type = habitType

	return updater.UpdateHabit(habit)
}

// ArchiveHabit устанавливает дату окончания привычки — она перестаёт отображаться в новых месяцах.
func (t *Tracker) ArchiveHabit(id int64) error {
	now := time.Now()
	return t.storage.HabitEndDate(id, now)
}

// DeleteHabit полностью удаляет привычку и все её записи из БД.
func (t *Tracker) DeleteHabit(id int64) error {
	return t.storage.DeleteHabit(id)
}

func (t *Tracker) GetEntries(date time.Time) ([]model.Entry, error) {
	entries, err := t.storage.GetEntries(date)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// GetEntryByCurrentDate получает entry по конкретной дате
func (t *Tracker) GetEntryByCurrentDate(habitID int64, date time.Time) (*model.Entry, error) {
	entries, err := t.storage.GetEntries(date)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.HabitID == habitID && sameDate(e.Date, date) {
			return &e, nil
		}
	}

	return nil, nil
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}

// AddEntry добавляет запись за конкретный день.
func (t *Tracker) AddEntry(habitID int64, date time.Time, value float64) (model.Entry, error) {
	e := model.Entry{
		HabitID: habitID,
		Date:    date,
		Value:   value,
	}

	err := t.storage.SaveEntry(e)
	if err != nil {
		return model.Entry{}, err
	}

	return e, nil
}

// UpdateEntry обновляет значение записи по id.
func (t *Tracker) UpdateEntry(id int64, value float64) (model.Entry, error) {
	return t.storage.UpdateEntry(id, value)
}

// DeleteEntry удаляет запись по id.
func (t *Tracker) DeleteEntry(id int64) error {
	return t.storage.DeleteEntry(id)
}
