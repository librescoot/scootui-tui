package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var blinkerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("214"))

// Blinker arrow shapes (3 lines tall)
var leftArrow = [3]string{
	"  █",
	"███",
	"  █",
}

var rightArrow = [3]string{
	"█  ",
	"███",
	"█  ",
}

// RenderBlinkers renders the blinker indicators.
// Returns 3 lines of output.
func RenderBlinkers(blinkerState string, flash bool, width int) string {
	if blinkerState == "off" || blinkerState == "" {
		return strings.Repeat("\n", 2)
	}

	totalWidth := max(width, 40)
	arrowWidth := 3

	var lines [3]string
	for i := 0; i < 3; i++ {
		left := strings.Repeat(" ", arrowWidth)
		right := strings.Repeat(" ", arrowWidth)

		if flash {
			if blinkerState == "left" || blinkerState == "both" {
				left = blinkerStyle.Render(leftArrow[i])
			}
			if blinkerState == "right" || blinkerState == "both" {
				right = blinkerStyle.Render(rightArrow[i])
			}
		}

		spacing := totalWidth - arrowWidth*2
		if spacing < 0 {
			spacing = 0
		}
		lines[i] = left + strings.Repeat(" ", spacing) + right
	}

	return strings.Join(lines[:], "\n")
}
