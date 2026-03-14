package main

import (
	"fmt"
	"scootui-tui/components"
	"scootui-tui/screens"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	screenTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	screenTabActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
)

// View implements tea.Model.
// Always fills exactly height lines: top frame, centered content, bottom frame.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	w := max(m.width, 60)
	h := max(m.height, 30)

	// Connection screen — fullscreen overlay
	if !m.redisClient.Connected {
		return m.renderFullscreenMessage(w, h, "Connecting to Redis...", m.redisClient.LastError)
	}

	// Vehicle state overlay for non-drivable states
	if msg, show := m.vehicleStateMessage(); show {
		return m.renderFullscreenMessage(w, h, msg, "")
	}

	// OTA overlay — fullscreen
	if m.ota != nil && m.ota.IsActive() {
		return m.renderOtaFullscreen(w, h)
	}

	// Normal screen rendering with frame
	return m.renderFramed(w, h)
}

// renderFramed renders the full screen: top bar (2 lines), content (centered), bottom bar (3 lines).
func (m Model) renderFramed(width, height int) string {
	sep := separatorStyle.Render(strings.Repeat("─", width))

	// Top frame: status bar + separator = 2 lines
	topBar := components.RenderTopStatusBar(
		m.battery0, m.battery1, m.vehicle, m.gps,
		m.bluetooth, m.internet, m.settings, width)

	// Bottom frame: separator + trip + battery + tabs = 4 lines
	tripBar := components.RenderBottomStatusBar(
		int(m.engine.SpeedKmh()), m.trip, m.engine.Odometer, width)

	batLine := components.RenderBatteryBar(m.battery0)
	if m.settings.DualBattery && m.battery1.Present {
		batLine += " " + components.RenderBatteryBar(m.battery1)
	}

	tabsLine := m.renderScreenTabs(width)

	// Content area: height - top(2) - bottom(4) = available lines
	contentHeight := height - 6
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Get screen content
	var content string
	switch m.activeScreen {
	case ScreenCluster:
		content = m.renderClusterContent(width, contentHeight)
	case ScreenNavigation:
		content = m.renderNavigationContent(width, contentHeight)
	case ScreenSettings:
		content = m.renderSettingsContent(width, contentHeight)
	case ScreenAbout:
		content = m.renderAboutContent(width, contentHeight)
	}

	// Vertically center the content within contentHeight
	contentLines := strings.Split(content, "\n")
	// Remove trailing empty line if present
	if len(contentLines) > 0 && contentLines[len(contentLines)-1] == "" {
		contentLines = contentLines[:len(contentLines)-1]
	}

	paddedContent := verticalCenter(contentLines, contentHeight, width)

	// Toast overlay — replaces top content line if active
	toastLine := components.RenderToast(&m.toasts, width)

	// Assemble full screen
	var b strings.Builder
	b.WriteString(topBar + "\n")
	if toastLine != "" {
		b.WriteString(toastLine + "\n")
	} else {
		b.WriteString(sep + "\n")
	}
	b.WriteString(paddedContent)
	b.WriteString(sep + "\n")
	b.WriteString(tripBar + "\n")
	b.WriteString(batLine + "\n")
	b.WriteString(tabsLine)

	return b.String()
}

// verticalCenter pads content lines to fill exactly `height` lines,
// centering the content vertically.
func verticalCenter(lines []string, height, width int) string {
	used := len(lines)
	if used >= height {
		// Content fills or exceeds available space — just join, truncate
		return strings.Join(lines[:min(used, height)], "\n") + "\n"
	}

	topPad := (height - used) / 2
	bottomPad := height - used - topPad

	var b strings.Builder
	for i := 0; i < topPad; i++ {
		b.WriteString("\n")
	}
	for _, line := range lines {
		b.WriteString(line + "\n")
	}
	for i := 0; i < bottomPad; i++ {
		b.WriteString("\n")
	}
	return b.String()
}

// Screen content renderers — return only the middle content, no top/bottom bars.

func (m Model) renderClusterContent(width, height int) string {
	return screens.RenderCluster(
		m.vehicle, m.engine, m.battery0, m.battery1,
		m.gps, m.bluetooth, m.internet, m.speedLimit,
		m.trip, m.settings, m.route, m.navigation,
		m.routeError, m.blinkerFlash, m.debugMode, width, height)
}

func (m Model) renderNavigationContent(width, height int) string {
	return screens.RenderNavigation(
		m.vehicle, m.engine, m.battery0, m.battery1,
		m.gps, m.bluetooth, m.internet, m.speedLimit,
		m.trip, m.settings, m.route, m.navigation,
		m.routeError, m.blinkerFlash, width, height)
}

func (m Model) renderSettingsContent(width, height int) string {
	return screens.RenderSettings(&m.menuState, width, height)
}

func (m Model) renderAboutContent(width, height int) string {
	return screens.RenderAbout(
		m.engine, m.battery0, m.battery1,
		m.internet, m.bluetooth, m.settings,
		m.aboutScroll, width, height)
}

// renderScreenTabs renders the bottom info line.
func (m Model) renderScreenTabs(width int) string {
	var left string

	switch m.activeScreen {
	case ScreenCluster:
		left = screenTabActiveStyle.Render("Cluster")
	case ScreenNavigation:
		left = screenTabActiveStyle.Render("Navigation")
	case ScreenSettings:
		left = screenTabActiveStyle.Render("Settings")
	case ScreenAbout:
		left = screenTabActiveStyle.Render("About")
	}

	hint := ""
	switch m.activeScreen {
	case ScreenCluster, ScreenNavigation:
		hint = "L:switch  L2x:menu"
	case ScreenSettings:
		hint = "L:scroll R:sel R-hold:back"
	case ScreenAbout:
		hint = "L:scroll R-hold:back"
	}

	hintStyled := screenTabStyle.Render(hint)

	leftWidth := lipgloss.Width(left)
	hintWidth := lipgloss.Width(hintStyled)
	gap := width - leftWidth - hintWidth
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + hintStyled
}

// Fullscreen overlays

func (m Model) renderFullscreenMessage(width, height int, title, detail string) string {
	center := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)

	var lines []string
	lines = append(lines, center.Render(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		Render(title)))
	if detail != "" {
		lines = append(lines, center.Render(lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(detail)))
	}

	return verticalCenter(lines, height, width)
}

func (m Model) vehicleStateMessage() (string, bool) {
	state := m.vehicle.State
	switch state {
	case "ready-to-drive", "stand-by", "parked", "":
		return "", false
	case "booting":
		return "Vehicle is booting...", true
	case "shutting-down":
		return "Vehicle is shutting down...", true
	case "hibernating", "hibernating-imminent":
		return "Vehicle is entering sleep mode...", true
	case "suspending", "suspending-imminent":
		return "Vehicle is suspending...", true
	case "updating":
		return "Vehicle is updating...", true
	default:
		return fmt.Sprintf("Vehicle state: %s", state), true
	}
}

func (m Model) renderOtaFullscreen(width, height int) string {
	otaContent := components.RenderOtaOverlay(m.ota, width)
	lines := strings.Split(otaContent, "\n")
	return verticalCenter(lines, height, width)
}
