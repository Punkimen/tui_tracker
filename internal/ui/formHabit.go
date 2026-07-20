package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"daily-tracker/internal/model"
	"daily-tracker/internal/tracker"
)

type (
	FormHabitModel struct {
		t           *tracker.Tracker
		habits      []model.Habit
		app         AppModel
		cursorHabit int
	}
)

type habitsData struct {
	habits []model.Habit
}

func (m AppModel) loadHabits() tea.Cmd {
	return func() tea.Msg {
		habits, err := m.tracker.GetHabits(m.currentDate)
		if err != nil {
			return mainData{}
		}

		return habitsData{
			habits,
		}
	}
}

func (f FormHabitModel) Init() tea.Cmd {
	return f.app.loadHabits()
}

func (f FormHabitModel) gridColumns() int {
	if len(f.habits) == 0 {
		return 0
	}

	columns := 3 + len(f.habits)/16
	if columns > 10 {
		return 10
	}

	return columns
}

func (f FormHabitModel) clampCursor() FormHabitModel {
	if len(f.habits) == 0 {
		f.cursorHabit = 0
		return f
	}

	if f.cursorHabit < 0 {
		f.cursorHabit = 0
	}
	if f.cursorHabit >= len(f.habits) {
		f.cursorHabit = len(f.habits) - 1
	}

	return f
}

func (f FormHabitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case habitsData:
		f.habits = msg.habits
		f = f.clampCursor()
	case tea.KeyPressMsg:
		columns := f.gridColumns()

		switch msg.String() {
		case "enter":
			if len(f.habits) == 0 {
				return f, nil
			}

			currentHabbit := f.habits[f.cursorHabit]
			return ChoosenModel{
				prevScreen: f,
				title:      fmt.Sprintf("Что сделать с привычкой %v", currentHabbit.Name),
				buttons: []Button{
					{
						label: "Update Habit",
						onSelect: func() (tea.Model, tea.Cmd) {
							return FormModel{
								t:         f.t,
								app:       f.app,
								field:     currentHabbit.Name,
								habitID:   currentHabbit.ID,
								habitForm: &f,
								typeHabit: currentHabbit.Type,
								goal:      currentHabbit.Goal,
							}, f.app.loadHabits()
						},
					},
					{
						label: "Delete Habit",
						onSelect: func() (tea.Model, tea.Cmd) {
							err := f.t.DeleteHabit(currentHabbit.ID)
							if err != nil {
								return f, nil
							}
							return f, f.app.loadHabits()
						},
					},
					{
						label: "Finish Habit",
						onSelect: func() (tea.Model, tea.Cmd) {
							err := f.t.ArchiveHabit(currentHabbit.ID)
							if err != nil {
								return f, nil
							}

							return f, f.app.loadData()
						},
					},
				},
			}, nil
		case "q":
			return f, tea.Quit
		case "esc":
			return f.app, nil
		case "up", "k":
			if f.cursorHabit-columns >= 0 {
				f.cursorHabit -= columns
			}
			return f, nil
		case "down", "j":
			if f.cursorHabit+columns < len(f.habits) {
				f.cursorHabit += columns
			}
			return f, nil
		case "left", "h":
			if f.cursorHabit > 0 {
				f.cursorHabit -= 1
			}
			return f, nil
		case "right", "l":
			if f.cursorHabit < len(f.habits)-1 {
				f.cursorHabit += 1
			}
			return f, nil
		}
	}

	return f, nil
}

func (f FormHabitModel) habitButtonWidth() int {
	const (
		minWidth = 12
		maxWidth = 24
	)

	width := minWidth
	for _, habit := range f.habits {
		labelWidth := lipgloss.Width(habit.Name)
		if labelWidth > width {
			width = labelWidth
		}
	}

	if width > maxWidth {
		return maxWidth
	}

	return width
}

func truncateLabel(label string, width int) string {
	if lipgloss.Width(label) <= width {
		return label
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	var b strings.Builder
	for _, r := range label {
		next := b.String() + string(r)
		if lipgloss.Width(next)+3 > width {
			break
		}
		b.WriteRune(r)
	}

	b.WriteString("...")
	return b.String()
}

func (f FormHabitModel) renderHabitButton(label string, focused bool, width int) string {
	label = truncateLabel(label, width)

	if focused {
		return ButtonFocusStyle.Width(width).Render(label)
	}

	return ButtonSecondaryStyle.Width(width).Render(label)
}

func (f FormHabitModel) View() tea.View {
	var b strings.Builder

	if len(f.habits) == 0 {
		b.WriteString("No habits")
		b.WriteString("\n\n")
		b.WriteString(renderNavigationHints(
			navigationHint{"esc", "назад"},
			navigationHint{"q", "выйти"},
		))
		return tea.NewView(b.String())
	}

	columns := f.gridColumns()
	buttonWidth := f.habitButtonWidth()

	for rowStart := 0; rowStart < len(f.habits); rowStart += columns {
		rowEnd := rowStart + columns
		if rowEnd > len(f.habits) {
			rowEnd = len(f.habits)
		}

		row := make([]string, 0, rowEnd-rowStart)
		for i := rowStart; i < rowEnd; i++ {
			row = append(
				row,
				f.renderHabitButton(f.habits[i].Name, f.cursorHabit == i, buttonWidth),
			)
		}

		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderNavigationHints(
		navigationHint{"enter", "выбрать"},
		navigationHint{"h/left", "влево"},
		navigationHint{"l/right", "вправо"},
		navigationHint{"j/down", "вниз"},
		navigationHint{"k/up", "вверх"},
		navigationHint{"esc", "назад"},
		navigationHint{"q", "выйти"},
	))

	return tea.NewView(b.String())
}

func (f FormHabitModel) DeleteHabit() tea.Cmd {
	habit := f.habits[f.cursorHabit]
	err := f.t.DeleteHabit(habit.ID)
	if err != nil {
		return nil
	}

	return f.app.loadData()
}

func (f FormHabitModel) UpdateHabit() (tea.Model, tea.Cmd) {
	if len(f.habits) == 0 || f.cursorHabit < 0 || f.cursorHabit >= len(f.habits) {
		return f, nil
	}

	habit := f.habits[f.cursorHabit]
	return FormModel{
		t:         f.t,
		habitForm: &f,
		focus:     field,
		editField: true,
		field:     habit.Name,
		typeHabit: habit.Type,
		goal:      habit.Goal,
		habitID:   habit.ID,
	}, nil
}
