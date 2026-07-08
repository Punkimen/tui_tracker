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

	editTable         bool
	currentEntryValue string
	err               string
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

func (m AppModel) updateHabit() (tea.Model, tea.Cmd) {
	day := m.days[m.cursorCol]
	habitID := m.habits[m.cursorRow].ID
	currentEntryDate := time.Date(m.now.Year(), m.now.Month(), day, 0, 0, 0, 0, m.now.Location())

	value, err := strconv.ParseFloat(strings.TrimSpace(m.currentEntryValue), 64)
	if err != nil {
		m.err = "Error with value"
		m.editTable = false
		m.currentEntryValue = ""
		return m, nil
	}

	entry, err := m.tracker.GetEntryByCurrentDate(habitID, currentEntryDate)
	if err != nil {
		m.err = err.Error()
		m.editTable = false
		m.currentEntryValue = ""
		return m, nil
	}

	if entry != nil {
		if _, err := m.tracker.UpdateEntry(entry.ID, value); err != nil {
			m.err = err.Error()
			m.editTable = false
			m.currentEntryValue = ""
			return m, nil
		}
	} else if _, err := m.tracker.AddEntry(habitID, currentEntryDate, value); err != nil {
		m.err = err.Error()
		m.editTable = false
		m.currentEntryValue = ""
		return m, nil
	}

	m.err = ""
	m.editTable = false
	m.currentEntryValue = ""
	return m, m.loadData()
}

func isNumberInput(value string) bool {
	if value == "" || value == "." {
		return true
	}

	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func (m AppModel) updateField(msg tea.KeyPressMsg) AppModel {
	switch msg.String() {
	case "backspace":
		runes := []rune(m.currentEntryValue)
		if len(runes) > 0 {
			m.currentEntryValue = string(runes[:len(runes)-1])
		}
	default:
		key := msg.String()
		if len([]rune(key)) == 1 {
			next := m.currentEntryValue + key
			if isNumberInput(next) {
				m.currentEntryValue = next
			}
		}
	}

	return m
}

func (m AppModel) navigationUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.currentFocus == buttonFocus {
			m.currentFocus = tableFocus
		}

		if m.cursorRow < len(m.habits)-1 {
			m.cursorRow += 1
		}

		return m, nil
	case "up", "k":
		if m.cursorRow == 0 && m.currentFocus == tableFocus {
			m.currentFocus = buttonFocus
		}

		if m.cursorRow > 0 && m.currentFocus == tableFocus {
			m.cursorRow -= 1
		}
	case "left", "h":
		if m.cursorCol != 0 && m.currentFocus == tableFocus {
			m.cursorCol -= 1
		}

	case "right", "l":
		if m.cursorCol < len(m.days)-1 && m.currentFocus == tableFocus {
			m.cursorCol += 1
		}
		return m, nil
	case "enter":
		if m.currentFocus == buttonFocus {
			return FormModel{
				t:         m.tracker,
				app:       m,
				typeHabit: model.HabitProgress,
				focus:     field,
				editField: true,
			}, nil
		}
		if m.currentFocus == tableFocus {

			habitType := m.habits[m.cursorRow].Type
			if habitType == "count" {
				habitID := m.habits[m.cursorRow].ID
				day := m.days[m.cursorCol]

				currentEntryDate := time.Date(
					m.now.Year(),
					m.now.Month(),
					day,
					0,
					0,
					0,
					0,
					m.now.Location(),
				)

				entry, err := m.tracker.GetEntryByCurrentDate(habitID, currentEntryDate)
				if err != nil {
					m.err = err.Error()
					m.editTable = false
					m.currentEntryValue = ""
					return m, nil
				}
				if entry != nil {
					value := 1.0
					if entry.Value != 0 {
						value = 0
					}

					if _, err := m.tracker.UpdateEntry(entry.ID, value); err != nil {
						m.err = err.Error()
						m.editTable = false
						m.currentEntryValue = ""
						return m, nil
					}
				} else if _, err := m.tracker.AddEntry(habitID, currentEntryDate, 1); err != nil {
					m.err = err.Error()
					m.editTable = false
					m.currentEntryValue = ""
					return m, nil
				}
				return m, nil
			}
			if m.editTable {
				return m.updateHabit()
			} else {
				m.editTable = true
				m.currentEntryValue = ""
			}
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
		if m.editTable {
			switch msg.String() {
			case "enter":
				return m.updateHabit()
			case "esc":
				m.editTable = false
				m.currentEntryValue = ""
				return m, nil
			}
			return m.updateField(msg), nil
		} else {
			return m.navigationUpdate(msg)
		}
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

	for i, v := range m.habits {
		habitRow := make([]string, len(m.days)+1)
		habitRow[0] = v.Name

		for col, day := range m.days {
			currentDate := time.Date(
				m.now.Year(),
				m.now.Month(),
				day,
				0, 0, 0, 0,
				m.now.Location(),
			)

			entry, err := m.tracker.GetEntryByCurrentDate(v.ID, currentDate)
			if err != nil {
				habitRow[col+1] = "!"
				continue
			}

			if entry != nil && entry.Value != 0 {
				habitRow[col+1] = strconv.FormatFloat(entry.Value, 'f', -1, 64)
			}
		}

		if m.cursorRow == i {
			if m.editTable && m.currentFocus == tableFocus {
				habitRow[m.cursorCol+1] = m.currentEntryValue
			} else {
				habitRow[m.cursorCol+1] = "X"
			}
		}

		b.WriteString(rowTable(habitRow, width))
		b.WriteString(borderTable(width))
		// allEntry, _ := m.tracker.GetAllEntries()

		// entries := make([]string, len(allEntry))
		//
		// for i, v := range allEntry {
		// 	if v.Value != 0 {
		// 		entries[i] = strconv.FormatFloat(v.Value, 'f', 0, 64)
		// 	} else {
		// 		entries[i] = "puk"
		// 	}
		// }
		// b.WriteString(rowTable(entries, width))
	}

	// for i, v := range m.habits {
	// 	habitRow := make([]string, len(m.days)+1)
	// 	habitRow[0] = v.Name
	// 	currentDate := time.Date(m.now.Year(), m.now.Month(), i+1, 0, 0, 0, 0, m.now.Location())
	// 	entry, err := m.tracker.GetEntryByCurrentDate(v.ID, currentDate)
	// 	if err != nil {
	// 		habitRow[m.cursorCol+1] = "!"
	// 	} else if entry != nil && entry.Value != 0 {
	// 		habitRow[m.cursorCol+1] = strconv.FormatFloat(entry.Value, 'f', 0, 64)
	// 	}
	// 	if m.cursorRow == i {
	// 		habitRow[m.cursorCol+1] = "X"
	// 	}
	// 	b.WriteString(rowTable(habitRow, width))
	// 	b.WriteString(borderTable(width))
	// }
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
	b.WriteString(fmt.Sprintf("\n %v", m.err))

	return tea.NewView(b.String())
}
