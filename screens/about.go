package screens

import (
	"fmt"
	"scootui-tui/redis"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	aboutTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	aboutLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	aboutValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	aboutWarnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	aboutDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// FOSS components used by scootui-tui
var fossComponents = [][2]string{
	{"Bubble Tea", "MIT"},
	{"Lip Gloss", "MIT"},
	{"go-redis", "BSD-2-Clause"},
	{"Go", "BSD-3-Clause"},
}

// RenderAbout renders the About & Licenses screen content, matching scootui.
func RenderAbout(
	engine *redis.EngineData,
	battery0, battery1 *redis.BatteryData,
	internet *redis.InternetData,
	bluetooth *redis.BluetoothData,
	settings *redis.SettingsData,
	scroll int,
	width, height int,
) string {
	var lines []string

	// Title
	lines = append(lines,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Render(aboutTitleStyle.Render("Librescoot")))
	lines = append(lines,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Render(aboutDimStyle.Render("ScootUI-TUI")))
	lines = append(lines, "")
	lines = append(lines,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Render(aboutDimStyle.Render("FOSS firmware for unu Scooter Pro")))
	lines = append(lines,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Render(aboutLabelStyle.Render("https://librescoot.org")))
	lines = append(lines, "")

	// License
	year := time.Now().Year()
	copyright := fmt.Sprintf("CC BY-NC-SA 4.0  2025-%d Librescoot", year)
	lines = append(lines,
		lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
			Render(aboutDimStyle.Render(copyright)))
	lines = append(lines, "")

	// Version info
	lines = append(lines, aboutTitleStyle.Render("VERSION"))
	lines = append(lines, infoLine("ECU Firmware", engine.FirmwareVersion, width))
	lines = append(lines, infoLine("Odometer", fmt.Sprintf("%.1f km", engine.Odometer/1000), width))
	lines = append(lines, "")

	// Non-commercial warning
	lines = append(lines, aboutWarnStyle.Render("NON-COMMERCIAL SOFTWARE"))
	lines = append(lines, wordWrap(
		"Commercial distribution, resale, or preinstallation "+
			"on devices for sale is prohibited under "+
			"CC BY-NC-SA 4.0.", width))
	lines = append(lines, "")
	lines = append(lines, wordWrap(
		"If you paid for this software or purchased a "+
			"scooter with it preinstalled, you may have been "+
			"scammed. Report at https://librescoot.org", width))
	lines = append(lines, "")

	// FOSS components
	lines = append(lines, aboutTitleStyle.Render("OPEN SOURCE COMPONENTS"))
	for _, comp := range fossComponents {
		name := comp[0]
		license := comp[1]
		// Two-column layout
		line := fmt.Sprintf("  %-28s %s", name, aboutDimStyle.Render(license))
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, aboutDimStyle.Render("L-brake: scroll  R-brake: exit"))

	// Apply scroll
	if scroll > len(lines)-height {
		scroll = len(lines) - height
	}
	if scroll < 0 {
		scroll = 0
	}

	end := scroll + height
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	if scroll > 0 {
		b.WriteString(aboutDimStyle.Render("  ^^^ scroll up ^^^"))
		b.WriteString("\n")
	}

	visible := lines[scroll:end]
	for _, line := range visible {
		b.WriteString(line)
		b.WriteString("\n")
	}

	if end < len(lines) {
		b.WriteString(aboutDimStyle.Render("  vvv scroll down vvv"))
	}

	return b.String()
}

func infoLine(label, value string, width int) string {
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("  %s  %s",
		aboutLabelStyle.Render(fmt.Sprintf("%-16s", label)),
		aboutValueStyle.Render(value))
}

func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	// Simple word wrap for the warning text
	words := strings.Fields(text)
	var lines []string
	current := "  " // indent
	for _, word := range words {
		if len(current)+len(word)+1 > width {
			lines = append(lines, current)
			current = "  " + word
		} else {
			if current == "  " {
				current += word
			} else {
				current += " " + word
			}
		}
	}
	if current != "  " {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}
