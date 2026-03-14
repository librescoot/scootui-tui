package components

import (
	"fmt"
	"scootui-tui/redis"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	connectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))

	disconnectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// RenderTopStatusBar renders the top status bar with battery, time, and connectivity.
func RenderTopStatusBar(
	battery0, battery1 *redis.BatteryData,
	vehicle *redis.VehicleData,
	gps *redis.GpsData,
	bluetooth *redis.BluetoothData,
	internet *redis.InternetData,
	settings *redis.SettingsData,
	width int,
) string {
	var left, right string

	// Left: Battery indicators
	if battery0.Present {
		left += RenderBatteryCompact(battery0)
		left += " "
	}
	if settings.DualBattery && battery1.Present {
		left += RenderBatteryCompact(battery1)
		left += " "
	}
	if vehicle.SeatboxLock == "open" {
		left += dimStyle.Render("[SEAT]") + " "
	}
	if battery0.Charge <= 20 || (settings.DualBattery && battery1.Charge <= 20) {
		left += warningStyle.Render("[SLOW]") + " "
	}

	// Center: Time
	center := time.Now().Format("15:04")

	// Right: Connectivity icons
	if settings.ShowGps != "never" {
		right += renderGpsIcon(gps, settings.ShowGps) + " "
	}
	if settings.ShowBluetooth != "never" {
		right += renderBluetoothIcon(bluetooth, settings.ShowBluetooth) + " "
	}
	if settings.ShowInternet != "never" {
		right += renderSignalBars(internet, settings.ShowInternet) + " "
	}

	// Layout
	totalWidth := max(width, 40)
	centerWidth := lipgloss.Width(center)
	sideWidth := (totalWidth - centerWidth - 2) / 2

	leftPadded := lipgloss.NewStyle().Width(sideWidth).Render(left)
	centerPadded := lipgloss.NewStyle().Width(centerWidth).Align(lipgloss.Center).Render(center)
	rightPadded := lipgloss.NewStyle().Width(sideWidth).Align(lipgloss.Right).Render(right)

	return leftPadded + " " + centerPadded + " " + rightPadded
}

// RenderBottomStatusBar renders the bottom trip statistics bar with current speed.
func RenderBottomStatusBar(speed int, trip *redis.TripData, odometer float64, width int) string {
	tripDist := fmt.Sprintf("%.1f", trip.DistanceKm())
	totalDist := fmt.Sprintf("%.1f", odometer/1000)

	speedStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).
		Render(fmt.Sprintf("%d km/h", speed))

	right := dimStyle.Render(fmt.Sprintf(
		"trip %s km  odo %s km", tripDist, totalDist))

	gap := width - lipgloss.Width(speedStyled) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return speedStyled + strings.Repeat(" ", gap) + right
}

func renderGpsIcon(gps *redis.GpsData, visibility string) string {
	switch gps.State {
	case "fix-established":
		if visibility == "error" {
			return ""
		}
		return connectedStyle.Render("GPS")
	case "searching":
		return warningStyle.Render("GPS?")
	case "error":
		return errorStyle.Render("GPS!")
	default:
		if visibility == "always" {
			return disconnectedStyle.Render("GPS-")
		}
		return ""
	}
}

func renderBluetoothIcon(bt *redis.BluetoothData, visibility string) string {
	if bt.Status == "connected" {
		if visibility == "error" {
			return ""
		}
		return connectedStyle.Render("BT")
	}
	if visibility == "always" {
		return disconnectedStyle.Render("BT-")
	}
	if visibility == "active-or-error" || visibility == "error" {
		return ""
	}
	return disconnectedStyle.Render("BT-")
}

func renderSignalBars(internet *redis.InternetData, visibility string) string {
	bars := internet.SignalBars()
	barChars := [4]string{"_", "_", "_", "_"}

	for i := 0; i < bars && i < 4; i++ {
		barChars[i] = string([]rune{'▂', '▄', '▆', '█'}[i])
	}

	display := strings.Join(barChars[:], "")

	if bars > 0 {
		if visibility == "error" {
			return ""
		}
		return connectedStyle.Render(display)
	}
	if visibility == "always" || visibility == "active-or-error" {
		return disconnectedStyle.Render(display)
	}
	return ""
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0:00"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d", hours, minutes)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
