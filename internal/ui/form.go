package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"daily-tracker/internal/model"
)

type formStep int

const (
	formStepName formStep = iota
	formStepType
)

// FormModel is a two-step form: enter name, then pick type.
// It is not a standalone tea.Model — AppModel drives it.
type FormModel struct {
	step      formStep
	name      string
	typeIdx   int
	confirmed bool
}

var habitTypes = []model.HabitType{
	model.HabitCount,
	model.HabitProgress,
	model.HabitMinutes,
}

func newFormModel() FormModel {
	return FormModel{}
}

func (f FormModel) selectedType() model.HabitType {
	return habitTypes[f.typeIdx]
}

// update returns (updated, done). done=true means the form is finished.
func (f FormModel) update(msg tea.Msg) (FormModel, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return f, false
	}

	if keyMsg.String() == "esc" {
		return f, true
	}

	switch f.step {
	case formStepName:
		switch keyMsg.String() {
		case "enter":
			if strings.TrimSpace(f.name) != "" {
				f.step = formStepType
			}
		case "backspace":
			if len(f.name) > 0 {
				f.name = f.name[:len(f.name)-1]
			}
		default:
			s := keyMsg.String()
			if len(s) == 1 {
				f.name += s
			}
		}

	case formStepType:
		switch keyMsg.String() {
		case "enter":
			f.confirmed = true
			return f, true
		case "left", "h":
			if f.typeIdx > 0 {
				f.typeIdx--
			}
		case "right", "l":
			if f.typeIdx < len(habitTypes)-1 {
				f.typeIdx++
			}
		}
	}

	return f, false
}

func (f FormModel) view() string {
	var b strings.Builder
	b.WriteString("\n  Create New Habit\n")
	b.WriteString("  ─────────────────\n\n")

	if f.step == formStepName {
		b.WriteString(fmt.Sprintf("  Name: %s_\n\n", f.name))
		b.WriteString("  Type the habit name and press Enter\n")
		b.WriteString("  Esc to cancel\n")
	} else {
		b.WriteString(fmt.Sprintf("  Name: %s\n\n", f.name))
		b.WriteString("  Type:\n  ")
		for i, t := range habitTypes {
			if i == f.typeIdx {
				b.WriteString(fmt.Sprintf("[ %s ]", t))
			} else {
				b.WriteString(fmt.Sprintf("  %s  ", t))
			}
			if i < len(habitTypes)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n\n")
		b.WriteString("  ←→ select type   Enter confirm   Esc cancel\n")
	}

	return b.String()
}
