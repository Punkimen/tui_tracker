package ui

import "github.com/charmbracelet/lipgloss"

var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#27DDF5")).
	MarginBottom(1)

var ButtonPrimaryStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#0B1020")).
	Background(lipgloss.Color("#27DDF5")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#27DDF5")).
	Padding(0, 2).
	MarginRight(1)

var ButtonSecondaryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#27DDF5")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#27DDF5")).
	Padding(0, 2).
	MarginRight(1)

var ButtonFocusStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#0B1020")).
	Background(lipgloss.Color("#F9E65C")).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#F9E65C")).
	Padding(0, 2).
	MarginRight(1)
