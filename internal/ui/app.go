// Package ui - TUI with bubletea
package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"daily-tracker/internal/model"
	"daily-tracker/internal/tracker"
)

type (
	screen int
	focus  int
)

const (
	appScreen screen = iota
	formScreen
)

const (
	buttonFocus focus = iota
	tableFocus
)

type AppModel struct {
	tracker *tracker.Tracker

	currentScreen screen
	currentFocus  focus

	cursorCol int
	cursorRow int

	habits []model.Habit
	entry  map[string]model.Entry

	now  time.Time
	days []int
}

func GetDaysFromMoth(month time.Month) []int {
	now := time.Now()
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()

	days := make([]int, lastDay)
	for i := range days {
		days[i] = i + 1
	}
	fmt.Printf("days %v, lastDay %v, month %v\n", days, lastDay, time.Month(month))
	return days
}

func CreateApp(t *tracker.Tracker) AppModel {
	// habits, err := t.GetActiveHabits(time.Now().Year(), time.Now().Month())
	now := time.Now()
	days := GetDaysFromMoth(now.Month())

	return AppModel{
		tracker:       t,
		currentScreen: appScreen,
		currentFocus:  buttonFocus,

		now:  now,
		days: days,
	}
}

type mainData struct {
	habits  []model.Habit
	entries []model.Entry
}

func (m AppModel) loadData() (mainData, error) {
	habits, err := m.tracker.GetActiveHabits(m.now.Year(), m.now.Month())
	if err != nil {
		return mainData{}, err
	}
	entries, err := m.tracker.GetAllEntries()
	if err != nil {
		return mainData{}, err
	}

	return mainData{
		habits, entries,
	}, nil
}

func (m AppModel) Init() tea.Cmd {
	return CreateApp
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
}
