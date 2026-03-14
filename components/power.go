package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	regenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))

	dischargeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	powerDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// RenderPowerBar renders a centered power bar showing discharge/regen.
// Negative power = regen (left), positive = discharge (right).
// maxPower is the scale reference in watts (e.g. 4000 for 4kW).
func RenderPowerBar(powerWatts float64, width int) string {
	barWidth := width - 20 // room for labels
	if barWidth < 10 {
		barWidth = 10
	}
	halfBar := barWidth / 2

	// Clamp to reasonable range
	maxW := 4000.0
	ratio := powerWatts / maxW
	if ratio > 1.0 {
		ratio = 1.0
	}
	if ratio < -1.0 {
		ratio = -1.0
	}

	filled := int(ratio * float64(halfBar))

	var leftSide, rightSide string

	if filled < 0 {
		// Regen: fill from center leftward
		regenBars := -filled
		leftSide = strings.Repeat("░", halfBar-regenBars) + regenStyle.Render(strings.Repeat("█", regenBars))
		rightSide = strings.Repeat("░", halfBar)
	} else if filled > 0 {
		// Discharge: fill from center rightward
		leftSide = strings.Repeat("░", halfBar)
		rightSide = dischargeStyle.Render(strings.Repeat("█", filled)) + strings.Repeat("░", halfBar-filled)
	} else {
		leftSide = strings.Repeat("░", halfBar)
		rightSide = strings.Repeat("░", halfBar)
	}

	// Labels
	var label string
	if powerWatts < -10 {
		label = regenStyle.Render(fmt.Sprintf("%.1fkW", powerWatts/1000))
	} else if powerWatts > 10 {
		label = dischargeStyle.Render(fmt.Sprintf("%.1fkW", powerWatts/1000))
	} else {
		label = powerDimStyle.Render("  0W")
	}

	bar := powerDimStyle.Render("RGN") + " " + leftSide + "|" + rightSide + " " + powerDimStyle.Render("PWR")

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(bar) +
		"\n" +
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(label)
}
