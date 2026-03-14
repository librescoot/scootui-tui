package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuItemType defines the behavior of a menu item.
type MenuItemType int

const (
	MenuSubmenu MenuItemType = iota
	MenuToggle
	MenuCycle
	MenuAction
)

// MenuItem represents a single menu entry.
type MenuItem struct {
	Label    string
	Key      string       // Redis settings key (for Toggle/Cycle)
	Type     MenuItemType
	Options  []string     // For Cycle type: available values
	Value    string       // Current value (for Toggle/Cycle display)
	Children []MenuItem   // For Submenu type
}

// MenuLevel represents one level of the menu hierarchy.
type MenuLevel struct {
	Title string
	Items []MenuItem
}

// MenuState holds the state of the hierarchical menu.
type MenuState struct {
	Stack   []MenuLevel
	Cursor  int
	Scroll  int // scroll offset for long menus
}

// NewMenuState creates a new menu state with the given root items.
func NewMenuState(title string, items []MenuItem) MenuState {
	return MenuState{
		Stack: []MenuLevel{{Title: title, Items: items}},
	}
}

// Current returns the current menu level.
func (m *MenuState) Current() *MenuLevel {
	if len(m.Stack) == 0 {
		return nil
	}
	return &m.Stack[len(m.Stack)-1]
}

// SelectedItem returns the currently selected menu item.
func (m *MenuState) SelectedItem() *MenuItem {
	level := m.Current()
	if level == nil || m.Cursor >= len(level.Items) {
		return nil
	}
	return &level.Items[m.Cursor]
}

// MoveUp moves the cursor up.
func (m *MenuState) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
		if m.Cursor < m.Scroll {
			m.Scroll = m.Cursor
		}
	}
}

// MoveDown moves the cursor down.
func (m *MenuState) MoveDown() {
	level := m.Current()
	if level != nil && m.Cursor < len(level.Items)-1 {
		m.Cursor++
	}
}

// Enter enters a submenu or activates the selected item.
// Returns the selected item if it's an action/toggle/cycle, nil if entering a submenu.
func (m *MenuState) Enter() *MenuItem {
	item := m.SelectedItem()
	if item == nil {
		return nil
	}

	if item.Type == MenuSubmenu && len(item.Children) > 0 {
		m.Stack = append(m.Stack, MenuLevel{
			Title: item.Label,
			Items: item.Children,
		})
		m.Cursor = 0
		m.Scroll = 0
		return nil
	}

	return item
}

// Back goes up one menu level. Returns false if already at root.
func (m *MenuState) Back() bool {
	if len(m.Stack) <= 1 {
		return false
	}
	m.Stack = m.Stack[:len(m.Stack)-1]
	m.Cursor = 0
	m.Scroll = 0
	return true
}

// Depth returns the current menu depth (1 = root).
func (m *MenuState) Depth() int {
	return len(m.Stack)
}

// RenderMenu renders the menu with cursor and scroll indicators.
func RenderMenu(state *MenuState, maxVisible int, width int) string {
	level := state.Current()
	if level == nil {
		return ""
	}

	var b strings.Builder

	// Breadcrumb
	var crumbs []string
	for _, l := range state.Stack {
		crumbs = append(crumbs, l.Title)
	}
	breadcrumb := strings.Join(crumbs, " > ")
	b.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render(breadcrumb))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(lipgloss.Width(breadcrumb)+4, width)))
	b.WriteString("\n")

	// Scroll state
	items := level.Items
	visibleStart := state.Scroll
	visibleEnd := visibleStart + maxVisible
	if visibleEnd > len(items) {
		visibleEnd = len(items)
	}

	// Adjust scroll to keep cursor visible
	if state.Cursor >= visibleEnd {
		state.Scroll = state.Cursor - maxVisible + 1
		visibleStart = state.Scroll
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(items) {
			visibleEnd = len(items)
		}
	}
	if state.Cursor < visibleStart {
		state.Scroll = state.Cursor
		visibleStart = state.Scroll
		visibleEnd = visibleStart + maxVisible
		if visibleEnd > len(items) {
			visibleEnd = len(items)
		}
	}

	// Scroll up indicator
	if visibleStart > 0 {
		b.WriteString(dimStyle.Render("  ^^^ more ^^^"))
		b.WriteString("\n")
	}

	// Menu items
	for i := visibleStart; i < visibleEnd; i++ {
		item := items[i]
		cursor := "  "
		if i == state.Cursor {
			cursor = "> "
		}

		label := item.Label
		value := ""

		switch item.Type {
		case MenuSubmenu:
			value = " >"
		case MenuToggle:
			if item.Value == "true" || item.Value == "on" {
				value = " [ON]"
			} else {
				value = " [OFF]"
			}
		case MenuCycle:
			if item.Value != "" {
				value = fmt.Sprintf(" [%s]", item.Value)
			}
		}

		line := cursor + label + value
		if i == state.Cursor {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Scroll down indicator
	if visibleEnd < len(items) {
		b.WriteString(dimStyle.Render("  vvv more vvv"))
		b.WriteString("\n")
	}

	// Controls hint
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("L-brake: scroll  R-brake: select  R-hold: back"))

	return b.String()
}

