package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

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
		habits, err := m.tracker.GetActiveHabits(m.now.Year(), m.now.Month())
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

func (f FormHabitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case habitsData:
		f.habits = msg.habits
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			currentHabbit := f.habits[f.cursorHabit]
			return ChoosenModel{
				prevScreen: f,
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
				},
			}, nil
		case "q":
			return f, tea.Quit
		case "esc":
			return f.app, nil
		case "up", "k":
			if f.cursorHabit > 0 {
				f.cursorHabit -= 1
			}
			return f, nil
		case "down", "j":
			if f.cursorHabit < len(f.habits)-1 {
				f.cursorHabit += 1
			}
			return f, nil
		}
	}

	return f, nil
}

func (f FormHabitModel) View() tea.View {
	var b strings.Builder

	for i, v := range f.habits {
		if f.cursorHabit == i {
			b.WriteString(fmt.Sprintf("[%v]\n", v.Name))
		} else {
			b.WriteString(fmt.Sprintf("%v \n", v.Name))
		}
	}

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
