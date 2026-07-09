package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type ChoosenModel struct {
	prevScreen tea.Model
	buttons    []Button
	focus      int
}

type Button struct {
	label    string
	onSelect func() (tea.Model, tea.Cmd)
}

func (c ChoosenModel) Init() tea.Cmd {
	return nil
}

func (c ChoosenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch key := msg.(type) {
	case tea.KeyPressMsg:
		switch key.String() {
		case "enter":
			currentButton := c.buttons[c.focus]
			return currentButton.onSelect()
		case "l", "right":
			if c.focus < len(c.buttons) {
				c.focus += 1
			}
		case "h", "left":
			if c.focus > 0 {
				c.focus -= 1
			}
		case "esc":
			return c.prevScreen, nil
		case "q":
			return c, tea.Quit
		}
	}
	return c, nil
}

func (c ChoosenModel) View() tea.View {
	var b strings.Builder
	for i, v := range c.buttons {
		if i == c.focus {
			b.WriteString(fmt.Sprintf("[%v] ", v.label))
		} else {
			b.WriteString(fmt.Sprintf("%v ", v.label))
		}
	}
	return tea.NewView(b.String())
}
