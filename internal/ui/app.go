// Package ui - TUI with bubletea
package ui

import (
	"fmt"
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
	entry  []model.Entry

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

func (m AppModel) loadData() tea.Cmd {
	return func() tea.Msg {
		habits, err := m.tracker.GetActiveHabits(m.now.Year(), m.now.Month())
		if err != nil {
			return mainData{}
		}
		entries, err := m.tracker.GetAllEntries()
		if err != nil {
			return mainData{}
		}

		return mainData{
			habits, entries,
		}
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadData()
}

func (m AppModel) navigationUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.currentFocus == buttonFocus {
			m.currentFocus = tableFocus
		}

		if m.cursorRow < len(m.habits)+1 {
			m.cursorRow += 1
		}

		return m, nil
	case "up", "h":
		if m.cursorRow == 0 && m.currentFocus == tableFocus {
			m.currentFocus = buttonFocus
		}

		if m.cursorRow > 0 && m.currentFocus == tableFocus {
			m.cursorRow -= 1
		}

		return m, nil
	case "enter":
		if m.currentFocus == buttonFocus {
			return FormModel{
				t:         m.tracker,
				app:       m,
				typeHabit: model.HabitProgress,
				focus:     field,
			}, nil
		}
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case mainData:
		m.habits = msg.habits
		m.entry = msg.entries
		return m, nil
	case tea.KeyPressMsg:
		return m.navigationUpdate(msg)
	}
	return m, nil
}

func borderTable(width []int) string {
	var b strings.Builder

	for _, i := range width {
		b.WriteByte('+')
		b.WriteString(strings.Repeat("-", i))
	}

	b.WriteByte('+')
	b.WriteString("\n")

	return b.String()
}

func rowTable(values []string, width []int) string {
	var b strings.Builder

	for i, v := range values {
		b.WriteByte('|')
		b.WriteString(fmt.Sprintf("%-*s", width[i], v))
	}
	b.WriteByte('|')
	b.WriteString("\n")

	return b.String()
}

func (m AppModel) renderTable(b *strings.Builder) {
	month := time.Time(m.now).Month().String()
	const (
		firstColWidth = 10
		colWidth      = 4
	)
	width := make([]int, len(m.days)+1)
	width[0] = firstColWidth
	for i := 1; i < len(width); i++ {
		width[i] = colWidth
	}

	row := make([]string, 0, len(m.days)+1)
	row = append(row, month)
	for _, d := range m.days {
		row = append(row, strconv.Itoa(d))
	}

	b.WriteString(borderTable(width))
	b.WriteString(rowTable(row, width))
	b.WriteString(borderTable(width))

	for _, v := range m.habits {
		habitRow := make([]string, len(m.days)+1)
		habitRow[0] = v.Name
		b.WriteString(rowTable(habitRow, width))
		b.WriteString(borderTable(width))
	}
}

func (m AppModel) View() tea.View {
	var b strings.Builder

	b.WriteString("Daily tracker\n")

	if m.currentFocus == buttonFocus {
		b.WriteString("[ Create habit ]\n")
	} else {
		b.WriteString("Create habit\n")
	}
	m.renderTable(&b)

	return tea.NewView(b.String())
}
