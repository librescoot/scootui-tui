package screens

import (
	"fmt"
	"scootui-tui/components"
	"scootui-tui/fonts"
	"scootui-tui/redis"
	"scootui-tui/valhalla"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	navTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	navDistStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	navInstrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	navSecondaryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	navSummaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	navIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	navErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	navDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// RenderNavigation renders the navigation screen content (no frame).
func RenderNavigation(
	vehicle *redis.VehicleData,
	engine *redis.EngineData,
	battery0, battery1 *redis.BatteryData,
	gps *redis.GpsData,
	bluetooth *redis.BluetoothData,
	internet *redis.InternetData,
	speedLimit *redis.SpeedLimitData,
	trip *redis.TripData,
	settings *redis.SettingsData,
	route *valhalla.Route,
	navigation *redis.NavigationData,
	routeError string,
	blinkerFlash bool,
	width, height int,
) string {
	var b strings.Builder

	// Compact blinkers
	if vehicle.BlinkerState != "off" && vehicle.BlinkerState != "" && blinkerFlash {
		b.WriteString(renderCompactBlinkers(string(vehicle.BlinkerState), width))
		b.WriteString("\n")
	}

	if route == nil {
		b.WriteString(renderNoRoute(navigation, routeError, width))
	} else {
		b.WriteString(renderActiveRoute(route, width, height))
	}

	// Auto-standby warning
	if warning := components.RenderAutoStandbyWarning(
		vehicle.AutoStandbyRemaining, width); warning != "" {
		b.WriteString("\n")
		b.WriteString(warning)
	}

	return b.String()
}

func renderNoRoute(navigation *redis.NavigationData, routeError string, width int) string {
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	var b strings.Builder

	if !navigation.HasDestination() {
		b.WriteString(center.Render(navSecondaryStyle.Render("No active navigation")))
		b.WriteString("\n")
		b.WriteString(center.Render(navDimStyle.Render("Set a destination via the app")))
	} else if routeError != "" {
		b.WriteString(center.Render(navErrorStyle.Render("Route error: " + routeError)))
	} else {
		b.WriteString(center.Render(navTitleStyle.Render("Calculating route...")))
	}
	b.WriteString("\n\n")

	return b.String()
}

func renderActiveRoute(route *valhalla.Route, width, height int) string {
	var b strings.Builder

	current := route.GetCurrentManeuver()
	if current == nil {
		return ""
	}

	// Large direction icon + instruction side by side
	iconType := fonts.ManeuverIcon(current.Type)
	iconLines := fonts.Icons[iconType]

	// Distance to the current maneuver's end (turn point)
	dist := valhalla.FormatDistance(current.Length)

	// Right-side text content
	var rightLines []string
	rightLines = append(rightLines, "")
	rightLines = append(rightLines, navDistStyle.Render("In "+dist))
	rightLines = append(rightLines, navInstrStyle.Render(current.Instruction))
	if street := current.GetStreetName(); street != "" {
		rightLines = append(rightLines, navSecondaryStyle.Render("on "+street))
	}

	next := route.GetNextManeuver()
	if next != nil && !route.IsComplete() {
		rightLines = append(rightLines, "")
		nextSmall := fonts.ManeuverSmallIcon(next.Type)
		rightLines = append(rightLines, navSecondaryStyle.Render(
			fmt.Sprintf("Then %s %s", nextSmall, next.Instruction)))
	}

	for len(rightLines) < fonts.IconHeight {
		rightLines = append(rightLines, "")
	}

	// Compose side by side
	iconWidth := fonts.IconWidth + 2
	rightWidth := width - iconWidth
	if rightWidth < 20 {
		rightWidth = 20
	}

	for i := 0; i < fonts.IconHeight; i++ {
		icon := ""
		if i < len(iconLines) {
			icon = navIconStyle.Render(iconLines[i])
		}
		iconPadded := lipgloss.NewStyle().Width(iconWidth).Render(icon)

		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		rightPadded := lipgloss.NewStyle().Width(rightWidth).Render(right)

		b.WriteString(iconPadded + rightPadded)
		b.WriteString("\n")
	}

	// Upcoming maneuvers list (use remaining height)
	upcoming := route.GetUpcomingManeuvers(6)
	if len(upcoming) > 1 {
		b.WriteString("\n")

		for i, m := range upcoming {
			if i == 0 {
				continue
			}
			icon := fonts.ManeuverSmallIcon(m.Type)
			dist := valhalla.FormatDistance(m.Length)

			instr := m.Instruction
			maxInstr := width - 10 // icon(2) + dist(~6) + spaces
			if maxInstr < 20 {
				maxInstr = 20
			}
			if len(instr) > maxInstr {
				instr = instr[:maxInstr-1] + "~"
			}

			line := fmt.Sprintf(" %s %-*s %s", icon, maxInstr, instr, dist)

			b.WriteString(navSecondaryStyle.Render(line))
			b.WriteString("\n")
		}
	}

	// Summary bar
	totalDist := valhalla.FormatDistance(route.TotalLength)
	totalTime := valhalla.FormatTime(route.TotalTime)
	b.WriteString("\n")
	b.WriteString(navSummaryStyle.Render(
		fmt.Sprintf("DST: %s remaining  |  ETA: %s", totalDist, totalTime)))
	b.WriteString("\n")

	return b.String()
}

func renderCompactBlinkers(state string, width int) string {
	left := "   "
	right := "   "

	if state == "left" || state == "both" {
		left = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("<<<")
	}
	if state == "right" || state == "both" {
		right = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(">>>")
	}

	spacing := width - 6
	if spacing < 0 {
		spacing = 0
	}
	return left + strings.Repeat(" ", spacing) + right
}
