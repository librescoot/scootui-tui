package input

import "time"

// Control identifies a physical control.
type Control int

const (
	ControlLeft  Control = iota // Left brake
	ControlRight               // Right brake
	ControlSeat                // Seatbox button
)

// Gesture identifies a detected gesture.
type Gesture int

const (
	GestureTap       Gesture = iota // Brief press (<500ms)
	GestureDoubleTap                // Two taps within 300ms
	GestureHold                     // Press held >500ms (fires once)
	GestureRelease                  // Released (after any press)
)

const (
	tapThreshold       = 500 * time.Millisecond
	doubleTapWindow    = 300 * time.Millisecond
	holdThreshold      = 500 * time.Millisecond
	seatDoublePressWin = 500 * time.Millisecond // seatbox uses wider window
)

// GestureEvent is emitted when a gesture is detected.
type GestureEvent struct {
	Control Control
	Gesture Gesture
}

// detector tracks state for one control.
type detector struct {
	pressed     bool
	pressTime   time.Time
	lastTapTime time.Time
	holding     bool
}

// GestureDetector detects tap, double-tap, and hold gestures from
// raw press/release events on physical controls.
type GestureDetector struct {
	detectors map[Control]*detector
	pending   []pendingTap
}

type pendingTap struct {
	control Control
	tapTime time.Time
}

func NewGestureDetector() *GestureDetector {
	return &GestureDetector{
		detectors: map[Control]*detector{
			ControlLeft:  {},
			ControlRight: {},
			ControlSeat:  {},
		},
	}
}

// Press records a button press event.
func (g *GestureDetector) Press(ctrl Control) []GestureEvent {
	d := g.detectors[ctrl]
	if d == nil {
		return nil
	}

	// Seatbox double-press detection (like Flutter's ShortcutMenuCubit):
	// check if this press is within the double-press window of the LAST press
	if ctrl == ControlSeat && !d.pressTime.IsZero() &&
		time.Since(d.pressTime) < seatDoublePressWin {
		d.pressed = true
		d.pressTime = time.Now()
		d.holding = false
		g.cancelPending(ctrl)
		return []GestureEvent{{Control: ctrl, Gesture: GestureDoubleTap}}
	}

	d.pressed = true
	d.pressTime = time.Now()
	d.holding = false
	return nil
}

// Release records a button release event. Returns detected gestures.
func (g *GestureDetector) Release(ctrl Control) []GestureEvent {
	d := g.detectors[ctrl]
	if d == nil {
		return nil
	}

	wasHolding := d.holding
	d.pressed = false
	d.holding = false

	var events []GestureEvent

	// Always emit Release (UI uses this for seatbox "release to confirm")
	events = append(events, GestureEvent{Control: ctrl, Gesture: GestureRelease})

	if wasHolding {
		// Was a hold — release already emitted above, nothing more to do
		return events
	}

	// Short press — potential tap
	duration := time.Since(d.pressTime)
	if duration >= tapThreshold {
		return events
	}

	// For brakes: check double-tap
	if ctrl != ControlSeat {
		if !d.lastTapTime.IsZero() && time.Since(d.lastTapTime) < doubleTapWindow {
			d.lastTapTime = time.Time{}
			g.cancelPending(ctrl)
			events = append(events, GestureEvent{Control: ctrl, Gesture: GestureDoubleTap})
			return events
		}
		// Defer tap to allow double-tap detection
		d.lastTapTime = time.Now()
		g.pending = append(g.pending, pendingTap{control: ctrl, tapTime: d.lastTapTime})
		return events
	}

	// Seatbox tap (double-press handled in Press())
	events = append(events, GestureEvent{Control: ctrl, Gesture: GestureTap})
	return events
}

// CheckHolds returns hold events for controls held past the threshold.
// Call every ~100ms.
func (g *GestureDetector) CheckHolds() []GestureEvent {
	var events []GestureEvent
	for ctrl, d := range g.detectors {
		if d.pressed && !d.holding && time.Since(d.pressTime) >= holdThreshold {
			d.holding = true
			events = append(events, GestureEvent{Control: ctrl, Gesture: GestureHold})
		}
	}
	return events
}

// FlushPending returns single taps whose double-tap window expired.
// Call every ~100ms.
func (g *GestureDetector) FlushPending() []GestureEvent {
	var events []GestureEvent
	var remaining []pendingTap

	now := time.Now()
	for _, p := range g.pending {
		if now.Sub(p.tapTime) >= doubleTapWindow {
			events = append(events, GestureEvent{Control: p.control, Gesture: GestureTap})
		} else {
			remaining = append(remaining, p)
		}
	}
	g.pending = remaining
	return events
}

// IsHolding returns whether the given control is currently held.
func (g *GestureDetector) IsHolding(ctrl Control) bool {
	d := g.detectors[ctrl]
	return d != nil && d.holding
}

func (g *GestureDetector) cancelPending(ctrl Control) {
	var remaining []pendingTap
	for _, p := range g.pending {
		if p.control != ctrl {
			remaining = append(remaining, p)
		}
	}
	g.pending = remaining
}

// ParseButtonEvent parses a Redis PUBSUB button event.
// Format: "brake:left:on", "seatbox:on", etc.
func ParseButtonEvent(event string) (Control, bool, bool) {
	switch event {
	case "brake:left:on":
		return ControlLeft, true, true
	case "brake:left:off":
		return ControlLeft, false, true
	case "brake:right:on":
		return ControlRight, true, true
	case "brake:right:off":
		return ControlRight, false, true
	case "seatbox:on":
		return ControlSeat, true, true
	case "seatbox:off":
		return ControlSeat, false, true
	default:
		return 0, false, false
	}
}
