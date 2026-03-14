package components

import (
	"scootui-tui/fonts"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	speedBlue = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	speedYellow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220"))

	speedRed = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	speedUnitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// RenderSpeedometer renders the speed as large figlet-style digits.
// Falls back to plain text if width < 40.
func RenderSpeedometer(speed int, width int) string {
	if width < 40 {
		return renderSpeedPlain(speed, width)
	}

	style := speedBlue
	if speed > 60 {
		style = speedRed
	} else if speed >= 55 {
		style = speedYellow
	}

	lines := fonts.RenderNumber(speed)
	var b strings.Builder

	for _, line := range lines {
		styled := style.Render(line)
		centered := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(styled)
		b.WriteString(centered)
		b.WriteString("\n")
	}

	unit := speedUnitStyle.Render("km/h")
	b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(unit))

	return b.String()
}

func renderSpeedPlain(speed int, width int) string {
	style := speedBlue
	if speed > 60 {
		style = speedRed
	} else if speed >= 55 {
		style = speedYellow
	}

	var b strings.Builder
	speedStr := style.Render(strings.Repeat(" ", 2) + lipgloss.NewStyle().Bold(true).Render(strings.TrimSpace(style.Render(formatInt(speed)))))
	b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(speedStr))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(speedUnitStyle.Render("km/h")))
	return b.String()
}

func formatInt(n int) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for v := n; v > 0; v /= 10 {
		s = string(rune('0'+v%10)) + s
	}
	return s
}
