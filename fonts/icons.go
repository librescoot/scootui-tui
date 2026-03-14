package fonts

// Large ASCII art direction icons for navigation.
// Each icon is 7 lines tall. Designed for linux framebuffer console.

type DirectionIcon int

const (
	IconStraight DirectionIcon = iota
	IconSlightRight
	IconRight
	IconSharpRight
	IconSlightLeft
	IconLeft
	IconSharpLeft
	IconUTurn
	IconRoundabout
	IconArrive
	IconMerge
	IconStart
)

const IconHeight = 7
const IconWidth = 12

var Icons = map[DirectionIcon][IconHeight]string{
	IconStraight: {
		"     ██     ",
		"    ████    ",
		"   ██████   ",
		"  ██ ██ ██  ",
		"     ██     ",
		"     ██     ",
		"     ██     ",
	},
	IconSlightRight: {
		"        ██  ",
		"      ████  ",
		"    ██████  ",
		"      ██    ",
		"     ██     ",
		"    ██      ",
		"    ██      ",
	},
	IconRight: {
		"     ██     ",
		"      ██    ",
		"████████████",
		"████████████",
		"      ██    ",
		"     ██     ",
		"            ",
	},
	IconSharpRight: {
		"    ██      ",
		"    ██      ",
		"    ██      ",
		"     ██     ",
		"      ████  ",
		"    ██████  ",
		"        ██  ",
	},
	IconSlightLeft: {
		"  ██        ",
		"  ████      ",
		"  ██████    ",
		"    ██      ",
		"     ██     ",
		"      ██    ",
		"      ██    ",
	},
	IconLeft: {
		"     ██     ",
		"    ██      ",
		"████████████",
		"████████████",
		"    ██      ",
		"     ██     ",
		"            ",
	},
	IconSharpLeft: {
		"      ██    ",
		"      ██    ",
		"      ██    ",
		"     ██     ",
		"  ████      ",
		"  ██████    ",
		"  ██        ",
	},
	IconUTurn: {
		"  ████████  ",
		" ██      ██ ",
		" ██      ██ ",
		"████     ██ ",
		"██████   ██ ",
		"████     ██ ",
		"         ██ ",
	},
	IconRoundabout: {
		"  ████████  ",
		" ██      ██ ",
		"██        ██",
		"██        ██",
		" ██    █████",
		"  ████████  ",
		"     ██     ",
	},
	IconArrive: {
		"  ████████  ",
		" ██      ██ ",
		" ██  ██  ██ ",
		" ██      ██ ",
		" ██  ██  ██ ",
		" ██      ██ ",
		"  ████████  ",
	},
	IconMerge: {
		"     ██     ",
		"     ██     ",
		"     ██     ",
		"    ████    ",
		"   ██  ▒▒   ",
		"  ██    ▒▒  ",
		" ██      ▒▒ ",
	},
	IconStart: {
		"     ██     ",
		"    ████    ",
		"   ██████   ",
		"  ██ ██ ██  ",
		"     ██     ",
		"     ██     ",
		"     ██     ",
	},
}

// ManeuverIcon maps Valhalla maneuver type codes to direction icons.
func ManeuverIcon(maneuverType int) DirectionIcon {
	switch maneuverType {
	case 1: // Start
		return IconStart
	case 2: // StartRight
		return IconSlightRight
	case 3: // StartLeft
		return IconSlightLeft
	case 4: // Destination
		return IconArrive
	case 5: // DestinationRight
		return IconArrive
	case 6: // DestinationLeft
		return IconArrive
	case 7, 8: // Becomes, Continue
		return IconStraight
	case 9: // SlightRight
		return IconSlightRight
	case 10: // Right
		return IconRight
	case 11: // SharpRight
		return IconSharpRight
	case 12, 13: // UTurnRight, UTurnLeft
		return IconUTurn
	case 14: // SharpLeft
		return IconSharpLeft
	case 15: // Left
		return IconLeft
	case 16: // SlightLeft
		return IconSlightLeft
	case 17, 22: // RampStraight, StayStraight
		return IconStraight
	case 18, 20, 23: // RampRight, ExitRight, StayRight
		return IconSlightRight
	case 19, 21, 24: // RampLeft, ExitLeft, StayLeft
		return IconSlightLeft
	case 25: // Merge
		return IconMerge
	case 26, 27: // RoundaboutEnter, RoundaboutExit
		return IconRoundabout
	default:
		return IconStraight
	}
}

// ManeuverSmallIcon returns a single-character icon for compact maneuver lists.
// Uses only characters safe for linux console.
func ManeuverSmallIcon(maneuverType int) string {
	switch maneuverType {
	case 1:
		return "^" // start
	case 4, 5, 6:
		return "X" // destination
	case 7, 8, 17, 22:
		return "|" // straight
	case 9, 18, 20, 23:
		return "/" // slight right
	case 10:
		return ">" // right
	case 11:
		return "}" // sharp right
	case 12, 13:
		return "U" // u-turn
	case 14:
		return "{" // sharp left
	case 15:
		return "<" // left
	case 16, 19, 21, 24:
		return `\` // slight left
	case 25:
		return "Y" // merge
	case 26, 27:
		return "O" // roundabout
	default:
		return ">"
	}
}
