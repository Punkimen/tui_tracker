// Package tracker implements business logic for the habit tracker.
package tracker

import (
	"fmt"
	"time"

	"daily-tracker/internal/model"
	"daily-tracker/internal/storage"
)

// Tracker — основная структура бизнес-логики.
// Хранит ссылку на storage через интерфейс, не через конкретный тип.
type Tracker struct {
	storage storage.Storage
}

// New — конструктор Tracker. Принимает любую реализацию Storage.
// Аналог в JS: function createTracker(storage: Storage) { return { storage } }
func New(s storage.Storage) *Tracker {
	return &Tracker{storage: s}
}

// AddHabit создаёт новую привычку.
// time.Now() проставляем здесь — это бизнес-логика, не дело storage.
func (t *Tracker) AddHabit(name string, habitType model.HabitType, goal *float64) (model.Habit, error) {
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
func (t *Tracker) GetHabits() ([]model.Habit, error) {
	return t.storage.GetHabits()
}

// GetActiveHabits возвращает только привычки активные в указанном месяце.
// Фильтрация происходит здесь, а не в SQL — логика принадлежит tracker, не storage.
func (t *Tracker) GetActiveHabits(year int, month time.Month) ([]model.Habit, error) {
	habits, err := t.storage.GetHabits()
	if err != nil {
		return nil, err
	}

	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)

	var active []model.Habit
	for _, h := range habits {
		if h.StartDate.After(last) {
			continue
		}
		if h.EndDate != nil && h.EndDate.Before(first) {
			continue
		}
		active = append(active, h)
	}

	return active, nil
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

// GetEntriesMap returns entries for a given month as a map keyed by "habitID-YYYY-MM-DD".
func (t *Tracker) GetEntriesMap(year int, month time.Month) (map[string]float64, error) {
	entries, err := t.storage.GetEntriesByMonth(year, month)
	if err != nil {
		return nil, err
	}
	m := make(map[string]float64, len(entries))
	for _, e := range entries {
		key := fmt.Sprintf("%d-%04d-%02d-%02d", e.HabitID, e.Date.Year(), int(e.Date.Month()), e.Date.Day())
		m[key] = e.Value
	}
	return m, nil
}

// UpsertEntry creates or updates an entry for a habit on a given date.
func (t *Tracker) UpsertEntry(habitID int64, date time.Time, value float64) error {
	return t.storage.UpsertEntry(habitID, date, value)
}
