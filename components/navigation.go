package components

import (
	"fmt"
	"scootui-tui/valhalla"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	navTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	navInstructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255"))

	navDistanceStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	navSecondaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	navSummaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// RenderNavigation renders the navigation display
func RenderNavigation(route *valhalla.Route, width int) string {
	if route == nil || len(route.Maneuvers) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	// Title
	b.WriteString(navTitleStyle.Render("NAVIGATION:"))
	b.WriteString("\n")

	// Primary instruction
	current := route.GetCurrentManeuver()
	if current != nil {
		icon := current.GetIcon()
		distance := valhalla.FormatDistance(route.RemainingDist)
		instruction := current.Instruction

		line := fmt.Sprintf("  %s In %s, %s",
			icon,
			navDistanceStyle.Render(distance),
			navInstructionStyle.Render(instruction))
		b.WriteString(line)
		b.WriteString("\n")

		// Secondary instruction (next maneuver)
		next := route.GetNextManeuver()
		if next != nil && !route.IsComplete() {
			secondary := fmt.Sprintf("     Then %s", next.Instruction)
			b.WriteString(navSecondaryStyle.Render(secondary))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Upcoming turns list
	upcoming := route.GetUpcomingManeuvers(4)
	if len(upcoming) > 0 {
		b.WriteString(navSecondaryStyle.Render("  Upcoming turns:"))
		b.WriteString("\n")

		for i, m := range upcoming {
			icon := m.GetIcon()
			street := m.GetStreetName()
			if street == "" {
				street = "unnamed road"
			}
			dist := valhalla.FormatDistance(m.Length)

			line := fmt.Sprintf("  %d. %s %s - %s (%s)",
				i+1, icon, m.Instruction, street, dist)
			b.WriteString(navSecondaryStyle.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Summary bar
	totalDist := valhalla.FormatDistance(route.TotalLength)
	totalTime := valhalla.FormatTime(route.TotalTime)

	summary := fmt.Sprintf("  DST: %s remaining  |  ETA: %s  |  arriving soon",
		totalDist, totalTime)
	b.WriteString(navSummaryStyle.Render(summary))
	b.WriteString("\n")

	return b.String()
}

// RenderNoNavigation renders a message when no navigation is active
func RenderNoNavigation() string {
	return navSecondaryStyle.Render("  No active navigation")
}
