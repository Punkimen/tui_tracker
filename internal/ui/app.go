// Package ui - TUI with bubletea
package ui

import (
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

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
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	key := keyMsg.String()
	switch key {
	case "enter":
		return m, nil

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

func (m AppModel) View() tea.View {
	var b strings.Builder

	b.WriteString("Daily tracker\n")

	if m.currentFocus == buttonFocus {
		b.WriteString("[ Create habit ]\n")
	} else {
		b.WriteString("Create habit\n")
	}

	for _, day := range m.days {
		b.WriteString(strconv.Itoa(day))
	}

	return tea.NewView(b.String())
}
