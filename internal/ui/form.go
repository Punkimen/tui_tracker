package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"daily-tracker/internal/model"
	"daily-tracker/internal/tracker"
)

type FormModel struct {
	t         *tracker.Tracker
	app       AppModel
	field     string
	typeHabit model.HabitType
	focus     int
	editField bool
	goal      *float64
	err       string
	habitID   int64
	habitForm *FormHabitModel
}

const (
	field       int = iota // 0
	progress               // 1
	count                  // 2
	minutes                // 3
	createHabit            // 4
)

func (f FormModel) Init() tea.Cmd {
	return nil
}

func (f FormModel) selectedTypeFocus() int {
	switch f.typeHabit {
	case model.HabitCount:
		return count
	case model.HabitMinutes:
		return minutes
	default:
		return progress
	}
}

func (f FormModel) setTypeFocus(focus int) FormModel {
	f.focus = focus

	switch focus {
	case count:
		f.typeHabit = model.HabitCount
	case minutes:
		f.typeHabit = model.HabitMinutes
	default:
		f.typeHabit = model.HabitProgress
	}

	return f
}

func (f FormModel) updateField(msg tea.KeyPressMsg) FormModel {
	switch msg.String() {
	case "backspace":
		runes := []rune(f.field)
		if len(runes) > 0 {
			f.field = string(runes[:len(runes)-1])
		}
	case "space":
		f.field += " "
	default:
		key := msg.String()
		if len([]rune(key)) == 1 {
			f.field += key
		}
	}

	return f
}

func (f FormModel) navigationUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if f.focus == field && f.editField && msg.String() != "enter" && msg.String() != "esc" {
		return f.updateField(msg), nil
	}

	switch msg.String() {
	case "left":
		switch f.focus {
		case count:
			f = f.setTypeFocus(progress)
		case minutes:
			f = f.setTypeFocus(count)
		}
	case "right":
		switch f.focus {
		case progress:
			f = f.setTypeFocus(count)
		case count:
			f = f.setTypeFocus(minutes)
		}
	case "up":
		switch f.focus {
		case progress, count, minutes:
			f.focus = field
			f.editField = false
		case createHabit:
			f.focus = f.selectedTypeFocus()
		}
	case "down":
		switch f.focus {
		case progress, count, minutes:
			f.focus = createHabit
		}
	case "esc":
		switch f.focus {
		case field:
			if f.editField {
				f.editField = false
			} else {
				if f.habitForm != nil {
					return f.habitForm, f.app.loadData()
				}
				return f.app, nil
			}
		case createHabit:
			f.focus = f.selectedTypeFocus()
		}
	case "enter":
		switch f.focus {
		case field:
			if !f.editField {
				f.editField = true
				f.err = ""
				return f, nil
			}

			if strings.TrimSpace(f.field) == "" {
				f.err = "Field is empty"
			} else {
				f.editField = false
				f.focus = f.selectedTypeFocus()
				f.err = ""
			}
		case progress, count, minutes:
			f.focus = createHabit
		case createHabit:
			if strings.TrimSpace(f.field) == "" {
				f.err = "Field is empty"
				f.focus = field
				f.editField = true
				return f, nil
			}

			if f.t == nil {
				f.err = "Tracker is not initialized"
				return f, nil
			}
			if f.habitID != 0 {
				if _, err := f.t.UpdateHabit(f.habitID, f.field, f.typeHabit); err != nil {
					f.err = err.Error()
					return f, nil
				}
			} else {
				if _, err := f.t.AddHabit(f.field, f.typeHabit, f.goal); err != nil {
					f.err = err.Error()
					return f, nil
				}
			}

			return f.app, f.app.loadData()
		}
	case "q":
		return f, tea.Quit

	}
	return f, nil
}

func (f FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return f.navigationUpdate(msg)
	}
	return f, nil
}

func (f FormModel) renderField(b *strings.Builder) {
	fieldValue := "Имя привычки: " + f.field

	if f.focus == field {
		b.WriteByte('>')
		b.WriteString(FieldFocusStyle.Render(fieldValue))
	} else {
		b.WriteByte(' ')
		b.WriteString(FieldStyle.Render(fieldValue))
	}

	b.WriteString("\n\n")
}

func (f FormModel) renderFormButton(label string, focused bool, selected bool) string {
	switch {
	case focused:
		return ButtonFocusStyle.Render(label)
	case selected:
		return ButtonPrimaryStyle.Render(label)
	default:
		return ButtonSecondaryStyle.Render(label)
	}
}

func (f FormModel) renderTypes(b *strings.Builder) {
	types := []struct {
		focus int
		label string
	}{
		{progress, "progress"},
		{count, "count"},
		{minutes, "minutes"},
	}

	buttons := make([]string, 0, len(types))
	for _, t := range types {
		selected := f.typeHabit ==
			model.HabitType(t.label)
		focused := f.focus == t.focus

		buttons = append(buttons, f.renderFormButton(t.label, focused, selected))
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, buttons...))
	b.WriteString("\n\n")
}

func (f FormModel) View() tea.View {
	var b strings.Builder

	f.renderField(&b)
	b.WriteString("\n")
	f.renderTypes(&b)
	b.WriteString(f.renderFormButton("Create Habit", f.focus == createHabit, false))

	if f.err != "" {
		b.WriteString("\n")
		b.WriteString(f.err)
	}

	b.WriteString("\n\n")
	if f.focus == field && f.editField {
		b.WriteString(renderNavigationHints(
			navigationHint{"enter", "подтвердить"},
			navigationHint{"esc", "отменить"},
		))
	} else {
		b.WriteString(renderNavigationHints(
			navigationHint{"enter", "выбрать"},
			navigationHint{"left", "влево"},
			navigationHint{"right", "вправо"},
			navigationHint{"down", "вниз"},
			navigationHint{"up", "вверх"},
			navigationHint{"esc", "назад"},
			navigationHint{"q", "выйти"},
		))
	}

	return tea.NewView(b.String())
}
