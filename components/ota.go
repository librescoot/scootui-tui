package components

import (
	"fmt"
	"scootui-tui/redis"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var otaStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("214"))

// RenderOtaOverlay renders the OTA update progress display.
// Returns empty string if no update is in progress.
func RenderOtaOverlay(ota *redis.OtaData, width int) string {
	if ota == nil || !ota.IsActive() {
		return ""
	}

	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("214")).
		Render("SYSTEM UPDATE IN PROGRESS"))
	b.WriteString("\n\n")

	if ota.MdbStatus != "" && ota.MdbStatus != "idle" {
		b.WriteString(renderOtaLine("MDB", ota.MdbStatus, ota.MdbUpdateVersion,
			ota.MdbDownloadProgress, ota.MdbInstallProgress, ota.MdbError, width))
		b.WriteString("\n")
	}

	if ota.DbcStatus != "" && ota.DbcStatus != "idle" {
		b.WriteString(renderOtaLine("DBC", ota.DbcStatus, ota.DbcUpdateVersion,
			ota.DbcDownloadProgress, ota.DbcInstallProgress, ota.DbcError, width))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("196")).
		Render("Do not power off the vehicle."))

	return b.String()
}

func renderOtaLine(component, status, version string, download, install int, otaErr string, width int) string {
	if otaErr != "" {
		return errorStyle.Render(fmt.Sprintf("  %s: Error - %s", component, otaErr))
	}

	var progress int
	var phase string
	switch status {
	case "downloading":
		progress = download
		phase = "Downloading"
	case "installing":
		progress = install
		phase = "Installing"
	default:
		phase = status
	}

	if version != "" {
		phase = fmt.Sprintf("%s %s", phase, version)
	}

	bar := renderProgressBar(progress, 20)
	return fmt.Sprintf("  %s: %s  %s %d%%", component, phase, bar, progress)
}

func renderProgressBar(percent, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := percent * width / 100
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
