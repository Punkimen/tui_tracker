package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"daily-tracker/internal/model"
	"daily-tracker/internal/tracker"
)

type (
	habitFormFocus int
	FormHabitModel struct {
		t            *tracker.Tracker
		habits       []model.Habit
		isEdit       bool
		field        string
		currentFocus habitFormFocus
		app          AppModel
		cursorHabit  int
	}
)

const (
	habits habitFormFocus = iota
	buttons
	form
)

func (f FormHabitModel) Init() tea.Cmd {
	return nil
}

func (f FormHabitModel) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return f, nil
}

func (f FormHabitModel) View() tea.View {
	var b strings.Builder

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
		habitForm: f,
		focus:     field,
		editField: true,
		field:     habit.Name,
		typeHabit: habit.Type,
		goal:      habit.Goal,
		habitID:   int(habit.ID),
	}, nil
}
