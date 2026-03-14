package components

import (
	"fmt"
	"scootui-tui/redis"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	batteryLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	batteryMidStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	batteryOkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("40"))
)

// RenderBatteryBar renders a battery as a visual bar: B0:[████████░░] 80%
func RenderBatteryBar(bat *redis.BatteryData) string {
	if !bat.Present {
		return ""
	}

	style := batteryOkStyle
	if bat.Charge <= 10 {
		style = batteryLowStyle
	} else if bat.Charge <= 20 {
		style = batteryMidStyle
	}

	if len(bat.Faults) > 0 {
		return errorStyle.Render(fmt.Sprintf("B%d:[!!FAULT!!] %d%%", bat.ID, bat.Charge))
	}

	barWidth := 10
	filled := bat.Charge * barWidth / 100
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	return fmt.Sprintf("B%d:[%s] %s",
		bat.ID,
		style.Render(bar),
		style.Render(fmt.Sprintf("%d%%", bat.Charge)),
	)
}

// RenderBatteryCompact renders a compact battery indicator for the status bar.
func RenderBatteryCompact(bat *redis.BatteryData) string {
	if !bat.Present {
		return ""
	}

	style := batteryOkStyle
	if bat.Charge <= 10 {
		style = batteryLowStyle
	} else if bat.Charge <= 20 {
		style = batteryMidStyle
	}

	if len(bat.Faults) > 0 {
		return errorStyle.Render(fmt.Sprintf("B%d:FAULT", bat.ID))
	}

	return style.Render(fmt.Sprintf("B%d:%d%%", bat.ID, bat.Charge))
}
