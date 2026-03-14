package screens

import (
	"fmt"
	"scootui-tui/components"
	"scootui-tui/redis"
	"scootui-tui/valhalla"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	navPreviewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	blinkerOnStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)

// RenderCluster renders the cluster screen content (no top/bottom bars).
// The caller handles the frame and vertical centering.
func RenderCluster(
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
	debugMode bool,
	width, height int,
) string {
	var b strings.Builder

	// Auto-standby warning banner (conditional)
	if warning := components.RenderAutoStandbyWarning(
		vehicle.AutoStandbyRemaining, width); warning != "" {
		b.WriteString(warning)
		b.WriteString("\n")
	}

	// Blinkers — single line
	blinkerLine := renderClusterBlinkers(string(vehicle.BlinkerState), blinkerFlash, width)
	if blinkerLine != "" {
		b.WriteString(blinkerLine)
		b.WriteString("\n")
	}

	// Speedometer (figlet, 5 lines + km/h label)
	b.WriteString(components.RenderSpeedometer(int(engine.SpeedKmh()), width))
	b.WriteString("\n")

	// Power bar
	power := engine.PowerOutput()
	b.WriteString(components.RenderPowerBar(power, width))
	b.WriteString("\n")

	// Status indicators (1 line)
	b.WriteString(components.RenderStatusIndicators(
		vehicle, engine, battery0, battery1, speedLimit, width))

	// Mini navigation preview (conditional)
	if route != nil {
		b.WriteString("\n")
		b.WriteString(renderMiniNav(route, width))
	} else if navigation.HasDestination() {
		b.WriteString("\n")
		if routeError != "" {
			b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
				Render(errorStyle.Render("Route: " + routeError)))
		} else {
			b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
				Render(navPreviewStyle.Render("Calculating route...")))
		}
	}

	// Debug info (conditional, uses remaining space)
	if debugMode {
		b.WriteString("\n")
		b.WriteString(renderDebugInfo(vehicle, engine, gps, battery0, battery1, route, width))
	}

	return b.String()
}

func renderClusterBlinkers(state string, flash bool, width int) string {
	if state == "off" || state == "" {
		return ""
	}

	left := "   "
	right := "   "

	if flash {
		if state == "left" || state == "both" {
			left = blinkerOnStyle.Render("<<<")
		}
		if state == "right" || state == "both" {
			right = blinkerOnStyle.Render(">>>")
		}
	}

	spacing := width - 6
	if spacing < 0 {
		spacing = 0
	}
	return left + strings.Repeat(" ", spacing) + right
}

func renderMiniNav(route *valhalla.Route, width int) string {
	current := route.GetCurrentManeuver()
	if current == nil {
		return ""
	}

	icon := current.GetIcon()
	dist := valhalla.FormatDistance(current.Length)
	line := fmt.Sprintf("%s In %s: %s", icon, dist, current.Instruction)

	next := route.GetNextManeuver()
	if next != nil && !route.IsComplete() {
		line += dimStyle.Render(fmt.Sprintf("  then %s", next.Instruction))
	}

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
		Render(navPreviewStyle.Render(line))
}

func renderDebugInfo(
	vehicle *redis.VehicleData,
	engine *redis.EngineData,
	gps *redis.GpsData,
	battery0, battery1 *redis.BatteryData,
	route *valhalla.Route,
	width int,
) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("DEBUG:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Veh:%s  Eng:%s\n", vehicle.State, engine.PowerState))
	b.WriteString(fmt.Sprintf("  GPS:%s %.4f,%.4f\n", gps.State, gps.Latitude, gps.Longitude))
	b.WriteString(fmt.Sprintf("  %dmV/%dmA %.0fC\n", engine.MotorVoltage, engine.MotorCurrent, engine.Temperature))
	b.WriteString(fmt.Sprintf("  B0:%d%% B1:%d%%", battery0.Charge, battery1.Charge))
	if route != nil {
		b.WriteString(fmt.Sprintf("\n  Route:%d/%d %.1fkm",
			route.CurrentManeuver, len(route.Maneuvers), route.RemainingDist))
	}
	return b.String()
}
