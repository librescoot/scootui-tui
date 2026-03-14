package fonts

import "strings"

// 7-line tall digit glyphs using block characters.
// Each digit is 8 chars wide.
var Digits = [10][7]string{
	// 0
	{
		" ██████ ",
		"██    ██",
		"██    ██",
		"██    ██",
		"██    ██",
		"██    ██",
		" ██████ ",
	},
	// 1
	{
		"   ██   ",
		"  ███   ",
		" ████   ",
		"   ██   ",
		"   ██   ",
		"   ██   ",
		" ██████ ",
	},
	// 2
	{
		" ██████ ",
		"██    ██",
		"      ██",
		"  ████  ",
		"██      ",
		"██      ",
		"████████",
	},
	// 3
	{
		" ██████ ",
		"██    ██",
		"      ██",
		"  █████ ",
		"      ██",
		"██    ██",
		" ██████ ",
	},
	// 4
	{
		"██    ██",
		"██    ██",
		"██    ██",
		"████████",
		"      ██",
		"      ██",
		"      ██",
	},
	// 5
	{
		"████████",
		"██      ",
		"██      ",
		"███████ ",
		"      ██",
		"██    ██",
		" ██████ ",
	},
	// 6
	{
		" ██████ ",
		"██      ",
		"██      ",
		"███████ ",
		"██    ██",
		"██    ██",
		" ██████ ",
	},
	// 7
	{
		"████████",
		"      ██",
		"     ██ ",
		"    ██  ",
		"   ██   ",
		"   ██   ",
		"   ██   ",
	},
	// 8
	{
		" ██████ ",
		"██    ██",
		"██    ██",
		" ██████ ",
		"██    ██",
		"██    ██",
		" ██████ ",
	},
	// 9
	{
		" ██████ ",
		"██    ██",
		"██    ██",
		" ███████",
		"      ██",
		"      ██",
		" ██████ ",
	},
}

const DigitHeight = 7
const DigitWidth = 8
const DigitSpacing = 1

// RenderNumber renders a number as large figlet-style text.
// Returns a slice of strings, one per line.
func RenderNumber(n int) []string {
	if n < 0 {
		n = 0
	}

	s := "0"
	if n > 0 {
		s = ""
		for v := n; v > 0; v /= 10 {
			s = string(rune('0'+v%10)) + s
		}
	}

	lines := make([]string, DigitHeight)
	for i := range lines {
		var parts []string
		for _, ch := range s {
			digit := int(ch - '0')
			parts = append(parts, Digits[digit][i])
		}
		lines[i] = strings.Join(parts, strings.Repeat(" ", DigitSpacing))
	}

	return lines
}

// RenderNumberWidth returns the total width of a rendered number.
func RenderNumberWidth(n int) int {
	digits := 1
	for v := n; v >= 10; v /= 10 {
		digits++
	}
	return digits*DigitWidth + (digits-1)*DigitSpacing
}
