package components

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ToastType determines the visual style.
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast represents a temporary notification message.
type Toast struct {
	Message string
	Type    ToastType
	Expires time.Time
}

var defaultDurations = map[ToastType]time.Duration{
	ToastInfo:    5 * time.Second,
	ToastSuccess: 2 * time.Second,
	ToastWarning: 3 * time.Second,
	ToastError:   15 * time.Second,
}

var toastStyles = map[ToastType]lipgloss.Style{
	ToastInfo: lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("24")),
	ToastSuccess: lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("22")),
	ToastWarning: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("214")),
	ToastError: lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("196")),
}

// ToastManager manages a stack of active toasts.
type ToastManager struct {
	toasts []Toast
}

// Show adds a toast with the default duration for its type.
func (tm *ToastManager) Show(msg string, t ToastType) {
	dur := defaultDurations[t]
	tm.toasts = append(tm.toasts, Toast{
		Message: msg,
		Type:    t,
		Expires: time.Now().Add(dur),
	})
}

// ShowFor adds a toast with a custom duration.
func (tm *ToastManager) ShowFor(msg string, t ToastType, dur time.Duration) {
	tm.toasts = append(tm.toasts, Toast{
		Message: msg,
		Type:    t,
		Expires: time.Now().Add(dur),
	})
}

// Prune removes expired toasts. Call periodically.
func (tm *ToastManager) Prune() {
	now := time.Now()
	var active []Toast
	for _, t := range tm.toasts {
		if t.Expires.After(now) {
			active = append(active, t)
		}
	}
	tm.toasts = active
}

// Active returns the most recent active toast, or nil.
func (tm *ToastManager) Active() *Toast {
	if len(tm.toasts) == 0 {
		return nil
	}
	return &tm.toasts[len(tm.toasts)-1]
}

// RenderToast renders the active toast as a full-width bar.
// Returns empty string if no active toast.
func RenderToast(tm *ToastManager, width int) string {
	t := tm.Active()
	if t == nil {
		return ""
	}

	style := toastStyles[t.Type]
	// Truncate message to fit width
	msg := t.Message
	if len(msg) > width-2 {
		msg = msg[:width-3] + "~"
	}
	// Pad to full width
	return style.Width(width).Render(" " + msg)
}
