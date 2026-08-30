package input

import "time"

type Control int

const (
	ControlLeft  Control = iota // Left brake
	ControlRight                // Right brake
	ControlSeat                 // Seatbox button
)

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

type GestureEvent struct {
	Control Control
	Gesture Gesture
}

type detector struct {
	pressed     bool
	pressTime   time.Time
	lastTapTime time.Time
	holding     bool
}

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

// Seat double-press is recognized on press (not release) to match the UI contract.
func (g *GestureDetector) Press(ctrl Control) []GestureEvent {
	d := g.detectors[ctrl]
	if d == nil {
		return nil
	}

	// Seatbox double presses are recognized on press, not release.
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

// Release is always emitted because seatbox shortcuts confirm on release.
func (g *GestureDetector) Release(ctrl Control) []GestureEvent {
	d := g.detectors[ctrl]
	if d == nil {
		return nil
	}

	wasHolding := d.holding
	d.pressed = false
	d.holding = false

	var events []GestureEvent

	events = append(events, GestureEvent{Control: ctrl, Gesture: GestureRelease})

	if wasHolding {
		return events
	}

	duration := time.Since(d.pressTime)
	if duration >= tapThreshold {
		return events
	}

	if ctrl != ControlSeat {
		if !d.lastTapTime.IsZero() && time.Since(d.lastTapTime) < doubleTapWindow {
			d.lastTapTime = time.Time{}
			g.cancelPending(ctrl)
			events = append(events, GestureEvent{Control: ctrl, Gesture: GestureDoubleTap})
			return events
		}
		d.lastTapTime = time.Now()
		g.pending = append(g.pending, pendingTap{control: ctrl, tapTime: d.lastTapTime})
		return events
	}

	events = append(events, GestureEvent{Control: ctrl, Gesture: GestureTap})
	return events
}

// CheckHolds is tick-driven; callers must invoke it roughly every 100 ms.
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

// FlushPending emits brake taps only after their double-tap window expires.
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

// Redis PUBSUB button-event format: "brake:left:on", "seatbox:on", etc.
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
