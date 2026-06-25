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

type appView int

const (
	viewMain appView = iota
	viewForm
)

type appFocus int

const (
	focusButton appFocus = iota
	focusTable
)

const nameWidth = 16
const cellWidth = 4

// AppModel is the root Bubble Tea model.
type AppModel struct {
	tracker *tracker.Tracker
	now     time.Time
	days    []int

	view    appView
	focus   appFocus
	habits  []model.Habit
	entries map[string]float64 // "habitID-YYYY-MM-DD" -> value

	cursorRow int
	cursorCol int
	colOffset int
	width     int

	editing bool
	editBuf string

	form FormModel
	err  error
}

type habitsLoadedMsg struct {
	habits  []model.Habit
	entries map[string]float64
}

type errMsg struct{ err error }

func NewApp(t *tracker.Tracker) AppModel {
	now := time.Now()
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	days := make([]int, lastDay)
	for i := range days {
		days[i] = i + 1
	}

	cursorCol := now.Day() - 1
	colOffset := max(0, cursorCol-4)

	return AppModel{
		tracker:   t,
		now:       now,
		days:      days,
		entries:   make(map[string]float64),
		cursorCol: cursorCol,
		colOffset: colOffset,
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadData()
}

func (m AppModel) loadData() tea.Cmd {
	return func() tea.Msg {
		habits, err := m.tracker.GetActiveHabits(m.now.Year(), m.now.Month())
		if err != nil {
			return errMsg{err}
		}
		entries, err := m.tracker.GetEntriesMap(m.now.Year(), m.now.Month())
		if err != nil {
			return errMsg{err}
		}
		return habitsLoadedMsg{habits, entries}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case habitsLoadedMsg:
		m.habits = msg.habits
		m.entries = msg.entries
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	if m.view == viewForm {
		return m.updateForm(msg)
	}
	return m.updateMain(msg)
}

func (m AppModel) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if m.editing {
		return m.updateEditing(key)
	}

	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.focus == focusTable {
			if m.cursorRow > 0 {
				m.cursorRow--
			} else {
				m.focus = focusButton
			}
		}

	case "down", "j":
		if m.focus == focusButton {
			if len(m.habits) > 0 {
				m.focus = focusTable
				m.cursorRow = 0
			}
		} else if m.cursorRow < len(m.habits)-1 {
			m.cursorRow++
		}

	case "left", "h":
		if m.focus == focusTable && m.cursorCol > 0 {
			m.cursorCol--
			if m.cursorCol < m.colOffset {
				m.colOffset = m.cursorCol
			}
		}

	case "right", "l":
		if m.focus == focusTable && m.cursorCol < len(m.days)-1 {
			m.cursorCol++
			visible := m.visibleCols()
			if m.cursorCol >= m.colOffset+visible {
				m.colOffset = m.cursorCol - visible + 1
			}
		}

	case "enter":
		if m.focus == focusButton {
			m.view = viewForm
			m.form = newFormModel()
		} else if m.focus == focusTable && len(m.habits) > 0 {
			m.editing = true
			key := entryKey(m.habits[m.cursorRow].ID, m.now.Year(), int(m.now.Month()), m.days[m.cursorCol])
			if v, ok := m.entries[key]; ok {
				m.editBuf = formatValue(v)
			} else {
				m.editBuf = ""
			}
		}
	}

	return m, nil
}

func (m AppModel) updateEditing(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		m.editing = false
		if v, err := strconv.ParseFloat(m.editBuf, 64); err == nil && m.editBuf != "" {
			habit := m.habits[m.cursorRow]
			day := m.days[m.cursorCol]
			date := time.Date(m.now.Year(), m.now.Month(), day, 0, 0, 0, 0, time.Local)
			eKey := entryKey(habit.ID, m.now.Year(), int(m.now.Month()), day)
			m.entries[eKey] = v
			return m, m.saveEntry(habit.ID, date, v)
		}

	case "esc":
		m.editing = false
		m.editBuf = ""

	case "backspace":
		if len(m.editBuf) > 0 {
			m.editBuf = m.editBuf[:len(m.editBuf)-1]
		}

	default:
		s := key.String()
		if len(s) == 1 && (s[0] >= '0' && s[0] <= '9' || s[0] == '.') {
			m.editBuf += s
		}
	}

	return m, nil
}

func (m AppModel) saveEntry(habitID int64, date time.Time, value float64) tea.Cmd {
	return func() tea.Msg {
		if err := m.tracker.UpsertEntry(habitID, date, value); err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m AppModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, done := m.form.update(msg)
	m.form = updated

	if !done {
		return m, nil
	}

	m.view = viewMain
	if m.form.confirmed {
		return m, m.createHabit(m.form.name, m.form.selectedType())
	}
	return m, nil
}

func (m AppModel) createHabit(name string, habitType model.HabitType) tea.Cmd {
	return func() tea.Msg {
		if _, err := m.tracker.AddHabit(name, habitType, nil); err != nil {
			return errMsg{err}
		}
		habits, err := m.tracker.GetActiveHabits(m.now.Year(), m.now.Month())
		if err != nil {
			return errMsg{err}
		}
		entries, err := m.tracker.GetEntriesMap(m.now.Year(), m.now.Month())
		if err != nil {
			return errMsg{err}
		}
		return habitsLoadedMsg{habits, entries}
	}
}

func (m AppModel) View() tea.View {
	if m.view == viewForm {
		return tea.NewView(m.form.view())
	}
	return tea.NewView(m.renderMain())
}

func (m AppModel) renderMain() string {
	var b strings.Builder

	if m.focus == focusButton {
		b.WriteString(" [ Create Habit ] <Enter>\n")
	} else {
		b.WriteString(" [ Create Habit ]\n")
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(fmt.Sprintf(" Error: %v\n", m.err))
		return b.String()
	}

	b.WriteString(fmt.Sprintf(" %s %d\n\n", m.now.Month(), m.now.Year()))

	if len(m.habits) == 0 {
		b.WriteString(" No habits yet — press Enter on the button above to create one.\n")
	} else {
		m.renderTable(&b)
	}

	b.WriteString("\n")
	if m.focus == focusButton {
		b.WriteString(" ↓ move to table   q quit\n")
	} else if m.editing {
		b.WriteString(" Type value   Enter confirm   Esc cancel\n")
	} else {
		b.WriteString(" ↑↓ habit   ←→ day   Enter edit cell   ↑ to button   q quit\n")
	}

	return b.String()
}

func (m AppModel) visibleCols() int {
	w := m.width
	if w == 0 {
		w = 80
	}
	available := w - 2 - nameWidth
	if available < cellWidth {
		return 1
	}
	return available / cellWidth
}

func (m AppModel) renderTable(b *strings.Builder) {
	visible := m.visibleCols()
	end := min(m.colOffset+visible, len(m.days))
	visibleDays := m.days[m.colOffset:end]

	// Header row
	b.WriteString(fmt.Sprintf("  %-*s", nameWidth, "Habit"))
	for _, d := range visibleDays {
		b.WriteString(fmt.Sprintf("%-*d", cellWidth, d))
	}
	b.WriteString("\n")

	// Separator
	b.WriteString("  " + strings.Repeat("─", nameWidth+len(visibleDays)*cellWidth) + "\n")

	for rowIdx, habit := range m.habits {
		rowFocused := m.focus == focusTable && m.cursorRow == rowIdx

		prefix := "  "
		if rowFocused {
			prefix = "> "
		}

		name := habit.Name
		if len(name) > nameWidth-1 {
			name = name[:nameWidth-1]
		}
		b.WriteString(fmt.Sprintf("%s%-*s", prefix, nameWidth, name))

		for colIdx := m.colOffset; colIdx < end; colIdx++ {
			day := m.days[colIdx]
			eKey := entryKey(habit.ID, m.now.Year(), int(m.now.Month()), day)
			cellVal := ""
			if v, ok := m.entries[eKey]; ok {
				cellVal = formatValue(v)
			}

			if rowFocused && m.cursorCol == colIdx {
				if m.editing {
					buf := m.editBuf
					if len(buf) > cellWidth-2 {
						buf = buf[len(buf)-(cellWidth-2):]
					}
					b.WriteString(fmt.Sprintf("[%-*s]", cellWidth-2, buf))
				} else {
					b.WriteString(fmt.Sprintf("[%-*s]", cellWidth-2, cellVal))
				}
			} else {
				b.WriteString(fmt.Sprintf("%-*s", cellWidth, cellVal))
			}
		}
		b.WriteString("\n")
	}
}

func entryKey(habitID int64, year, month, day int) string {
	return fmt.Sprintf("%d-%04d-%02d-%02d", habitID, year, month, day)
}

func formatValue(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.1f", v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
