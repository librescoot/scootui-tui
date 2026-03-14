package screens

import (
	"scootui-tui/components"
	"strings"
)

// RenderSettings renders the settings screen with hierarchical menu.
func RenderSettings(menuState *components.MenuState, width, height int) string {
	var b strings.Builder

	// Calculate visible menu items based on terminal height
	// Reserve lines for: breadcrumb(2), controls(2), padding(2)
	maxVisible := height - 6
	if maxVisible < 5 {
		maxVisible = 5
	}
	if maxVisible > 20 {
		maxVisible = 20
	}

	b.WriteString("\n")
	b.WriteString(components.RenderMenu(menuState, maxVisible, width))
	b.WriteString("\n")

	return b.String()
}
