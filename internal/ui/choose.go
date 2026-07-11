package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type ChoosenModel struct {
	prevScreen tea.Model
	title      string
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
			if len(c.buttons) == 0 {
				return c.prevScreen, nil
			}

			currentButton := c.buttons[c.focus]
			return currentButton.onSelect()
		case "l", "right":
			if c.focus < len(c.buttons)-1 {
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

func placeOverlay(overlay string, background string) string {
	backgroundWidth := lipgloss.Width(background)
	backgroundHeight := lipgloss.Height(background)
	overlayHeight := lipgloss.Height(overlay)
	startRow := (backgroundHeight - overlayHeight) / 2
	if startRow < 0 {
		startRow = 0
	}

	backgroundLines := strings.Split(background, "\n")
	for len(backgroundLines) < startRow+overlayHeight {
		backgroundLines = append(backgroundLines, "")
	}

	overlayLines := strings.Split(overlay, "\n")
	for i, line := range overlayLines {
		backgroundLines[startRow+i] = lipgloss.PlaceHorizontal(
			backgroundWidth,
			lipgloss.Center,
			line,
		)
	}

	return strings.Join(backgroundLines, "\n")
}

func (c ChoosenModel) renderButton(index int, button Button) string {
	if index == c.focus {
		return ButtonFocusStyle.Render(button.label)
	}

	return ButtonSecondaryStyle.Render(button.label)
}

func (c ChoosenModel) renderModal() string {
	var b strings.Builder

	if c.title != "" {
		b.WriteString(ModalTitleStyle.Render(c.title))
		b.WriteString("\n")
	}

	buttons := make([]string, 0, len(c.buttons))
	for i, button := range c.buttons {
		buttons = append(buttons, c.renderButton(i, button))
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, buttons...))
	return ModalStyle.Render(b.String())
}

func (c ChoosenModel) View() tea.View {
	background := c.prevScreen.View().Content
	modal := c.renderModal()

	return tea.NewView(placeOverlay(modal, background))
}
