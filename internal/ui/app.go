// Package ui - TUI with bubletea
package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"daily-tracker/internal/model"
	"daily-tracker/internal/tracker"
)

type (
	screen int
	focus  int
)

type buttonType struct {
	label    string
	onSelect func(app AppModel) (tea.Model, tea.Cmd)
	style    lipgloss.Style
}

const (
	appScreen screen = iota
	formScreen
)

const (
	buttonsFocus focus = iota
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

	buttons        []buttonType
	buttonPosition int

	editTable         bool
	currentEntryValue string
	err               string

	windowWidth int
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

	app := AppModel{
		tracker:        t,
		currentScreen:  appScreen,
		currentFocus:   buttonsFocus,
		buttonPosition: 0,
		now:            now,
		days:           days,
	}
	app.InitButtons()
	return app
}

func (m *AppModel) InitButtons() {
	buttons := make([]buttonType, 2)
	buttons[0] = buttonType{
		label: "Create Habit",
		style: ButtonPrimaryStyle,
		onSelect: func(app AppModel) (tea.Model, tea.Cmd) {
			return FormModel{
				t:         app.tracker,
				app:       app,
				typeHabit: model.HabitProgress,
				focus:     field,
				editField: true,
			}, nil
		},
	}

	buttons[1] = buttonType{
		label: "Update Habits",
		style: ButtonSecondaryStyle,
		onSelect: func(app AppModel) (tea.Model, tea.Cmd) {
			return FormHabitModel{
				t:           app.tracker,
				app:         app,
				habits:      app.habits,
				cursorHabit: 0,
			}, nil
		},
	}
	m.buttons = buttons
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
		if m.currentFocus == buttonsFocus {
			m.currentFocus = tableFocus
		}

		if m.cursorRow < len(m.habits)-1 {
			m.cursorRow += 1
		}

		return m, nil
	case "up", "k":
		if m.cursorRow == 0 && m.currentFocus == tableFocus {
			m.currentFocus = buttonsFocus
		}

		if m.cursorRow > 0 && m.currentFocus == tableFocus {
			m.cursorRow -= 1
		}
	case "left", "h":
		if m.currentFocus == buttonsFocus && m.buttonPosition != 0 {
			m.buttonPosition -= 1
		}
		if m.cursorCol != 0 && m.currentFocus == tableFocus {
			m.cursorCol -= 1
		}

	case "right", "l":
		if m.currentFocus == buttonsFocus && m.buttonPosition < len(m.buttons)-1 {
			m.buttonPosition += 1
		}
		if m.cursorCol < len(m.days)-1 && m.currentFocus == tableFocus {
			m.cursorCol += 1
		}
		return m, nil
	case "enter":
		if m.currentFocus == buttonsFocus {
			currentButton := m.buttons[m.buttonPosition]
			return currentButton.onSelect(m)
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
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil
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
	}
}

func (m AppModel) renderTitle() string {
	const banner = `
            ██████╗  █████╗ ██╗██╗  ██╗   ██╗            
            ██╔══██╗██╔══██╗██║██║  ╚██╗ ██╔╝            
            ██║  ██║███████║██║██║   ╚████╔╝             
            ██║  ██║██╔══██║██║██║    ╚██╔╝              
            ██████╔╝██║  ██║██║███████╗██║               
            ╚═════╝ ╚═╝  ╚═╝╚═╝╚══════╝╚═╝               
                                                         
████████╗██████╗  █████╗  ██████╗██╗  ██╗███████╗██████╗ 
╚══██╔══╝██╔══██╗██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗
   ██║   ██████╔╝███████║██║     █████╔╝ █████╗  ██████╔╝
   ██║   ██╔══██╗██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
   ██║   ██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`
	title := TitleStyle.Render(strings.TrimPrefix(banner, "\n"))

	if m.windowWidth <= 0 {
		return title
	}

	return lipgloss.PlaceHorizontal(m.windowWidth, lipgloss.Center, title)
}

func (m AppModel) renderButton(index int, button buttonType) string {
	if m.currentFocus == buttonsFocus && index == m.buttonPosition {
		return ButtonFocusStyle.Render(button.label)
	}

	return button.style.Render(button.label)
}

func (m AppModel) renderButtons() string {
	buttons := make([]string, len(m.buttons))
	for i, v := range m.buttons {
		buttons[i] = m.renderButton(i, v)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, buttons...)
}

func (m AppModel) View() tea.View {
	var b strings.Builder

	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	b.WriteString(m.renderButtons())
	b.WriteString("\n")
	m.renderTable(&b)
	b.WriteString(fmt.Sprintf("\n %v", m.err))

	return tea.NewView(b.String())
}
