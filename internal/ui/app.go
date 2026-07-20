// Package ui - TUI with bubletea
package ui

import (
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

const (
	tableFirstColWidth = 15
	tableDayColWidth   = 6
)

type AppModel struct {
	tracker *tracker.Tracker

	currentScreen screen
	currentFocus  focus

	cursorCol int
	cursorRow int
	colOffset int
	rowOffset int

	habits []model.Habit
	entry  []model.Entry

	currentDate time.Time
	days        []int

	buttons        []buttonType
	buttonPosition int

	editTable         bool
	currentEntryValue string
	err               string

	windowWidth  int
	windowHeight int
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
		currentDate:    now,
		days:           days,
		cursorCol:      now.Day() - 1,
	}
	app.InitButtons()
	return app
}

func (m *AppModel) InitButtons() {
	buttons := make([]buttonType, 4)
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

	buttons[2] = buttonType{
		label: "←",
		style: ButtonSecondaryStyle,
		onSelect: func(app AppModel) (tea.Model, tea.Cmd) {
			app.currentDate = app.currentDate.AddDate(0, -1, 0)
			lastDay := time.Date(
				app.currentDate.Year(),
				app.currentDate.Month()+1,
				0, 0, 0, 0, 0,
				app.currentDate.Location(),
			).Day()
			app.days = make([]int, lastDay)
			for i := range app.days {
				app.days[i] = i + 1
			}

			return app, app.loadData()
		},
	}
	buttons[3] = buttonType{
		label: "→",
		style: ButtonSecondaryStyle,
		onSelect: func(app AppModel) (tea.Model, tea.Cmd) {
			app.currentDate = app.currentDate.AddDate(0, 1, 0)
			lastDay := time.Date(
				app.currentDate.Year(),
				app.currentDate.Month()+1,
				0, 0, 0, 0, 0,
				app.currentDate.Location(),
			).Day()
			app.days = make([]int, lastDay)
			for i := range app.days {
				app.days[i] = i + 1
			}

			return app, app.loadData()
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
		habits, err := m.tracker.GetHabits(m.currentDate)
		if err != nil {
			return mainData{}
		}
		entries, err := m.tracker.GetEntries(m.currentDate)
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
	habbitEndDate := m.habits[m.cursorRow].EndDate

	currentEntryDate := time.Date(
		m.currentDate.Year(),
		m.currentDate.Month(),
		day,
		0,
		0,
		0,
		0,
		m.currentDate.Location(),
	)

	if habbitEndDate != nil &&
		currentEntryDate.Format("2006-01-02") > habbitEndDate.Format("2006-01-02") {
		m.err = "You finish this Habit"
		return m, nil
	}

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

func (m AppModel) clampTableCursor() AppModel {
	if len(m.habits) == 0 {
		m.cursorRow = 0
		m.rowOffset = 0
	} else {
		if m.cursorRow < 0 {
			m.cursorRow = 0
		}
		if m.cursorRow >= len(m.habits) {
			m.cursorRow = len(m.habits) - 1
		}
	}

	if len(m.days) == 0 {
		m.cursorCol = 0
		m.colOffset = 0
	} else {
		if m.cursorCol < 0 {
			m.cursorCol = 0
		}
		if m.cursorCol >= len(m.days) {
			m.cursorCol = len(m.days) - 1
		}
	}

	return m
}

func (m AppModel) visibleDayCount() int {
	if len(m.days) == 0 {
		return 0
	}
	if m.windowWidth <= 0 {
		return len(m.days)
	}

	available := m.windowWidth - tableFirstColWidth - 2
	if available <= 0 {
		return 1
	}

	visible := available / (tableDayColWidth + 1)
	if visible < 1 {
		return 1
	}
	if visible > len(m.days) {
		return len(m.days)
	}

	return visible
}

func (m AppModel) visibleHabitCount() int {
	if len(m.habits) == 0 {
		return 0
	}
	if m.windowHeight <= 0 {
		return len(m.habits)
	}

	const (
		tableHeaderHeight = 3
		tableRowHeight    = 2
		viewSpacingHeight = 3
	)

	available := m.windowHeight -
		lipgloss.Height(m.renderTitle()) -
		lipgloss.Height(m.renderButtons()) -
		viewSpacingHeight -
		tableHeaderHeight

	if available <= 0 {
		return 1
	}

	visible := available / tableRowHeight
	if visible < 1 {
		return 1
	}
	if visible > len(m.habits) {
		return len(m.habits)
	}

	return visible
}

func (m AppModel) syncTableViewport() AppModel {
	m = m.clampTableCursor()

	visibleRows := m.visibleHabitCount()
	if visibleRows <= 0 {
		m.rowOffset = 0
	} else {
		if m.cursorRow < m.rowOffset {
			m.rowOffset = m.cursorRow
		}
		if m.cursorRow >= m.rowOffset+visibleRows {
			m.rowOffset = m.cursorRow - visibleRows + 1
		}
		maxOffset := len(m.habits) - visibleRows
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.rowOffset > maxOffset {
			m.rowOffset = maxOffset
		}
		if m.rowOffset < 0 {
			m.rowOffset = 0
		}
	}

	visibleCols := m.visibleDayCount()
	if visibleCols <= 0 {
		m.colOffset = 0
	} else {
		if m.cursorCol < m.colOffset {
			m.colOffset = m.cursorCol
		}
		if m.cursorCol >= m.colOffset+visibleCols {
			m.colOffset = m.cursorCol - visibleCols + 1
		}
		maxOffset := len(m.days) - visibleCols
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.colOffset > maxOffset {
			m.colOffset = maxOffset
		}
		if m.colOffset < 0 {
			m.colOffset = 0
		}
	}

	return m
}

func (m AppModel) navigationUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "down", "j":
		if m.cursorRow < len(m.habits)-1 && m.currentFocus == tableFocus {
			m.cursorRow += 1
		}

		if m.currentFocus == buttonsFocus && len(m.habits) > 0 {
			m.currentFocus = tableFocus
		}
		return m.syncTableViewport(), nil
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
		return m.syncTableViewport(), nil
	case "enter":
		if m.currentFocus == buttonsFocus {
			currentButton := m.buttons[m.buttonPosition]
			return currentButton.onSelect(m)
		}
		if m.currentFocus == tableFocus {
			if len(m.habits) == 0 || len(m.days) == 0 {
				return m, nil
			}

			habitType := m.habits[m.cursorRow].Type
			habbitEndDate := m.habits[m.cursorRow].EndDate
			day := m.days[m.cursorCol]

			currentEntryDate := time.Date(
				m.currentDate.Year(),
				m.currentDate.Month(),
				day,
				0,
				0,
				0,
				0,
				m.currentDate.Location(),
			)

			if habbitEndDate != nil &&
				currentEntryDate.Format("2006-01-02") > habbitEndDate.Format("2006-01-02") {
				m.err = "You finish this Habit"
				return m, nil
			}

			if habitType == "count" {
				habitID := m.habits[m.cursorRow].ID

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
	return m.syncTableViewport(), nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
		return m.syncTableViewport(), nil
	case mainData:
		m.habits = msg.habits
		m.entry = msg.entries
		return m.syncTableViewport(), nil
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
		b.WriteString(padCell(v, width[i]))
	}
	b.WriteByte('|')
	b.WriteString("\n")

	return b.String()
}

func padCell(value string, width int) string {
	valueWidth := lipgloss.Width(value)
	if valueWidth >= width {
		return value
	}

	return value + strings.Repeat(" ", width-valueWidth)
}

func focusedTableCell(value string, width int) string {
	return TableCellFocusStyle.Width(width).Render(value)
}

func visibleRange(offset int, visible int, total int) (int, int) {
	if total <= 0 || visible <= 0 {
		return 0, 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		offset = total - 1
	}

	end := offset + visible
	if end > total {
		end = total
	}

	return offset, end
}

func tableCellValue(value string, width int) string {
	return truncateLabel(value, width)
}

func (m AppModel) renderTable(b *strings.Builder) {
	// month := time.Time(m.currentDate).Month().String()
	dateString := time.Time(m.currentDate).Format("Jan 2006")
	visibleDays := m.visibleDayCount()
	visibleHabits := m.visibleHabitCount()
	dayStart, dayEnd := visibleRange(m.colOffset, visibleDays, len(m.days))
	habitStart, habitEnd := visibleRange(m.rowOffset, visibleHabits, len(m.habits))

	width := make([]int, dayEnd-dayStart+1)
	width[0] = tableFirstColWidth
	for i := 1; i < len(width); i++ {
		width[i] = tableDayColWidth
	}

	row := make([]string, 0, len(width))
	row = append(row, dateString)
	for _, d := range m.days[dayStart:dayEnd] {
		row = append(row, strconv.Itoa(d))
	}

	b.WriteString(borderTable(width))
	b.WriteString(rowTable(row, width))
	b.WriteString(borderTable(width))

	for i := habitStart; i < habitEnd; i++ {
		v := m.habits[i]
		habitRow := make([]string, len(width))
		habitValue := TableCellFocusStyle.Width(width[0]).Render(v.Name)
		habitEndDate := v.EndDate

		if m.cursorRow == i && m.currentFocus == tableFocus {
			habitValue = TableCellFocusStyle.Width(width[0]).Render(v.Name)
		} else {
			habitValue = v.Name
		}

		habitRow[0] = tableCellValue(habitValue, tableFirstColWidth)
		for col, day := range m.days[dayStart:dayEnd] {
			currentDate := time.Date(
				m.currentDate.Year(),
				m.currentDate.Month(),
				day,
				0, 0, 0, 0,
				m.currentDate.Location(),
			)

			if habitEndDate != nil &&
				currentDate.Format("2006-01-02") > habitEndDate.Format("2006-01-02") {
				habitRow[col+1] = TableCellDisabled.Width(tableDayColWidth).Render("")
			}

			entry, err := m.tracker.GetEntryByCurrentDate(v.ID, currentDate)
			if err != nil {
				habitRow[col+1] = "!"
				continue
			}

			if entry != nil && entry.Value != 0 {
				value := strconv.FormatFloat(entry.Value, 'f', -1, 64)
				habitRow[col+1] = tableCellValue(value, tableDayColWidth)
			}
		}

		if m.cursorRow == i && m.currentFocus == tableFocus {
			focusedCol := m.cursorCol - dayStart + 1
			focusedValue := habitRow[focusedCol]
			if m.editTable {
				focusedValue = m.currentEntryValue
			}
			habitRow[focusedCol] = focusedTableCell(
				tableCellValue(focusedValue, width[focusedCol]),
				width[focusedCol],
			)
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

	if m.err != "" {
		b.WriteString("\n ")
		b.WriteString(ErrorHintStyle.Render(m.err))
	}

	b.WriteString("\n\n")
	if m.editTable {
		b.WriteString(renderNavigationHints(
			navigationHint{"enter", "сохранить"},
			navigationHint{"esc", "отменить"},
		))
	} else {
		b.WriteString(renderNavigationHints(
			navigationHint{"enter", "выбрать"},
			navigationHint{"h/left", "влево"},
			navigationHint{"l/right", "вправо"},
			navigationHint{"j/down", "вниз"},
			navigationHint{"k/up", "вверх"},
			navigationHint{"q", "выйти"},
		))
	}

	return tea.NewView(b.String())
}
