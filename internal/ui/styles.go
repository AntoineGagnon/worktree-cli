package ui

import "github.com/charmbracelet/lipgloss"

var (
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	previewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	keyStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
)
