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

// View always fills the terminal height; safety/connection overlays replace the
// normal screen before framed content is rendered.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	w := max(m.width, 60)
	h := max(m.height, 30)

	if !m.redisClient.Connected {
		return m.renderFullscreenMessage(w, h, "Connecting to Redis...", m.redisClient.LastError)
	}

	if msg, show := m.vehicleStateMessage(); show {
		return m.renderFullscreenMessage(w, h, msg, "")
	}

	if m.ota != nil && m.ota.IsActive() {
		return m.renderOtaFullscreen(w, h)
	}

	return m.renderFramed(w, h)
}

// renderFramed reserves two top and four bottom lines before centering content.
func (m Model) renderFramed(width, height int) string {
	sep := separatorStyle.Render(strings.Repeat("─", width))

	topBar := components.RenderTopStatusBar(
		m.battery0, m.battery1, m.vehicle, m.gps,
		m.bluetooth, m.internet, m.settings, width)

	tripBar := components.RenderBottomStatusBar(
		int(m.engine.SpeedKmh()), m.trip, m.engine.Odometer, width)

	batLine := components.RenderBatteryBar(m.battery0)
	if m.settings.DualBattery && m.battery1.Present {
		batLine += " " + components.RenderBatteryBar(m.battery1)
	}

	tabsLine := m.renderScreenTabs(width)

	contentHeight := height - 6
	if contentHeight < 10 {
		contentHeight = 10
	}

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

	contentLines := strings.Split(content, "\n")
	if len(contentLines) > 0 && contentLines[len(contentLines)-1] == "" {
		contentLines = contentLines[:len(contentLines)-1]
	}

	paddedContent := verticalCenter(contentLines, contentHeight, width)

	toastLine := components.RenderToast(&m.toasts, width)

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

// verticalCenter truncates overflow but otherwise produces exactly height lines.
func verticalCenter(lines []string, height, width int) string {
	used := len(lines)
	if used >= height {
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
