package components

import (
	"fmt"
	"scootui-tui/redis"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusIndicators renders the priority-based status indicator line.
func RenderStatusIndicators(
	vehicle *redis.VehicleData,
	engine *redis.EngineData,
	battery0, battery1 *redis.BatteryData,
	speedLimit *redis.SpeedLimitData,
	width int,
) string {
	var indicator string

	if vehicle.IsUnableToDrive == "on" {
		indicator = warningStyle.Render("!! Unable to drive")
	} else if vehicle.BlinkerState == "both" {
		indicator = warningStyle.Render("!! Hazard lights on")
	} else if vehicle.BrakeLeft == "on" || vehicle.BrakeRight == "on" {
		indicator = warningStyle.Render("[P] Parking brake engaged")
	} else if len(battery0.Faults) > 0 || len(battery1.Faults) > 0 {
		indicator = errorStyle.Render("!! Battery fault")
	} else if vehicle.AutoStandbyRemaining > 0 && vehicle.AutoStandbyRemaining <= 30 {
		indicator = warningStyle.Render(fmt.Sprintf(
			"Auto-standby in %ds - press brake", vehicle.AutoStandbyRemaining))
	} else if speedLimit.HasSpeedLimit() {
		indicator = fmt.Sprintf("[%d km/h] %s", speedLimit.SpeedLimitInt(), speedLimit.RoadName)
	} else {
		power := engine.PowerOutput()
		if power < 0 {
			indicator = batteryOkStyle.Render(fmt.Sprintf("%.1f kW regen", -power/1000))
		} else if power > 0 {
			indicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Render(fmt.Sprintf("PWR %.1f kW", power/1000))
		}
	}

	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(indicator)
}

// RenderAutoStandbyWarning renders a prominent auto-standby warning banner.
func RenderAutoStandbyWarning(remaining int, width int) string {
	if remaining <= 0 || remaining > 30 {
		return ""
	}

	msg := fmt.Sprintf("!! AUTO-STANDBY IN %ds - PRESS BRAKE !!", remaining)

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("208")).
		Width(width).
		Align(lipgloss.Center)

	return style.Render(msg)
}
