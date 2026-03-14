package valhalla

import "fmt"

// RouteRequest represents a Valhalla routing request
type RouteRequest struct {
	Locations         []Location        `json:"locations"`
	Costing           string            `json:"costing"`
	DirectionsOptions DirectionsOptions `json:"directions_options"`
}

// Location represents a coordinate
type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// DirectionsOptions contains routing options
type DirectionsOptions struct {
	Units string `json:"units"`
}

// RouteResponse represents a Valhalla routing response
type RouteResponse struct {
	Trip Trip `json:"trip"`
}

// Trip contains route information
type Trip struct {
	Summary  Summary    `json:"summary"`
	Legs     []Leg      `json:"legs"`
	Status   int        `json:"status"`
	StatusMessage string `json:"status_message,omitempty"`
}

// Summary contains route summary
type Summary struct {
	Length float64 `json:"length"` // km
	Time   float64 `json:"time"`   // seconds
}

// Leg represents a route leg
type Leg struct {
	Summary   Summary    `json:"summary"`
	Maneuvers []Maneuver `json:"maneuvers"`
}

// Maneuver represents a turn instruction
type Maneuver struct {
	Type              int     `json:"type"`
	Instruction       string  `json:"instruction"`
	StreetNames       []string `json:"street_names,omitempty"`
	Length            float64 `json:"length"` // km
	Time              float64 `json:"time"`   // seconds
	BeginShapeIndex   int     `json:"begin_shape_index"`
	EndShapeIndex     int     `json:"end_shape_index"`
	RoundaboutExitCount int   `json:"roundabout_exit_count,omitempty"`
}

// Maneuver types (from Valhalla documentation)
const (
	ManeuverStart               = 1
	ManeuverStartRight          = 2
	ManeuverStartLeft           = 3
	ManeuverDestination         = 4
	ManeuverDestinationRight    = 5
	ManeuverDestinationLeft     = 6
	ManeuverBecomes             = 7
	ManeuverContinue            = 8
	ManeuverSlightRight         = 9
	ManeuverRight               = 10
	ManeuverSharpRight          = 11
	ManeuverUturnRight          = 12
	ManeuverUturnLeft           = 13
	ManeuverSharpLeft           = 14
	ManeuverLeft                = 15
	ManeuverSlightLeft          = 16
	ManeuverRampStraight        = 17
	ManeuverRampRight           = 18
	ManeuverRampLeft            = 19
	ManeuverExitRight           = 20
	ManeuverExitLeft            = 21
	ManeuverStayStraight        = 22
	ManeuverStayRight           = 23
	ManeuverStayLeft            = 24
	ManeuverMerge               = 25
	ManeuverRoundaboutEnter     = 26
	ManeuverRoundaboutExit      = 27
	ManeuverFerryEnter          = 28
	ManeuverFerryExit           = 29
	ManeuverTransit             = 30
	ManeuverTransitTransfer     = 31
	ManeuverTransitRemainOn     = 32
	ManeuverTransitConnectionStart = 33
	ManeuverTransitConnectionTransfer = 34
	ManeuverTransitConnectionDestination = 35
	ManeuverPostTransitConnectionDestination = 36
)

// GetIcon returns a terminal-safe icon for the maneuver type.
func (m *Maneuver) GetIcon() string {
	switch m.Type {
	case ManeuverStart, ManeuverContinue, ManeuverStayStraight:
		return "^"
	case ManeuverSlightRight:
		return "/"
	case ManeuverRight, ManeuverStartRight, ManeuverDestinationRight:
		return ">"
	case ManeuverSharpRight:
		return "}"
	case ManeuverSlightLeft:
		return `\`
	case ManeuverLeft, ManeuverStartLeft, ManeuverDestinationLeft:
		return "<"
	case ManeuverSharpLeft:
		return "{"
	case ManeuverUturnRight, ManeuverUturnLeft:
		return "U"
	case ManeuverRoundaboutEnter, ManeuverRoundaboutExit:
		return "O"
	case ManeuverDestination:
		return "X"
	case ManeuverMerge:
		return "Y"
	default:
		return ">"
	}
}

// GetStreetName returns the street name or empty string
func (m *Maneuver) GetStreetName() string {
	if len(m.StreetNames) > 0 {
		return m.StreetNames[0]
	}
	return ""
}

// FormatDistance formats distance for display
func FormatDistance(km float64) string {
	if km < 1.0 {
		return fmt.Sprintf("%dm", int(km*1000))
	}
	return fmt.Sprintf("%.1fkm", km)
}

// FormatTime formats time for display
func FormatTime(seconds float64) string {
	minutes := int(seconds / 60)
	if minutes < 60 {
		return fmt.Sprintf("%dmin", minutes)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh %dmin", hours, mins)
}

// Route represents a calculated route with maneuvers
type Route struct {
	StartLat        float64
	StartLon        float64
	EndLat          float64
	EndLon          float64
	TotalLength     float64 // km
	TotalTime       float64 // seconds
	Maneuvers       []Maneuver
	CurrentManeuver int
	RemainingDist   float64 // km to current maneuver
}

// GetCurrentManeuver returns the current maneuver
func (r *Route) GetCurrentManeuver() *Maneuver {
	if r.CurrentManeuver >= 0 && r.CurrentManeuver < len(r.Maneuvers) {
		return &r.Maneuvers[r.CurrentManeuver]
	}
	return nil
}

// GetNextManeuver returns the next maneuver
func (r *Route) GetNextManeuver() *Maneuver {
	next := r.CurrentManeuver + 1
	if next >= 0 && next < len(r.Maneuvers) {
		return &r.Maneuvers[next]
	}
	return nil
}

// GetUpcomingManeuvers returns next N maneuvers
func (r *Route) GetUpcomingManeuvers(n int) []Maneuver {
	start := r.CurrentManeuver
	if start < 0 {
		start = 0
	}
	end := start + n
	if end > len(r.Maneuvers) {
		end = len(r.Maneuvers)
	}
	return r.Maneuvers[start:end]
}

// IsComplete returns true if route is complete
func (r *Route) IsComplete() bool {
	return r.CurrentManeuver >= len(r.Maneuvers)-1
}
