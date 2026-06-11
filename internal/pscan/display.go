package pscan

import (
	"fmt"
	"strings"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
	Gray    = "\033[90m"
)

// ColorMem returns an ANSI color based on memory usage threshold.
func ColorMem(mbVal float64) string {
	switch {
	case mbVal > 500:
		return Red
	case mbVal > 200:
		return Yellow
	case mbVal > 50:
		return Cyan
	default:
		return Green
	}
}

// StateColor returns an ANSI color based on process state.
func StateColor(s string) string {
	switch s {
	case "Z":
		return Red
	case "S", "S+":
		return Green
	case "I":
		return Gray
	default:
		return Yellow
	}
}

// Colorize wraps text with color and reset.
func Colorize(text, color string) string {
	return color + text + Reset
}

// BoldText returns bold-styled text.
func BoldText(text string) string {
	return Bold + text + Reset
}

// DimText returns gray-styled text.
func DimText(text string) string {
	return Gray + text + Reset
}

// Header prints a section header line.
func Header(format string, args ...interface{}) {
	fmt.Printf("\n  "+Bold+format+Reset+"\n\n", args...)
}

// TableHeader prints a table header with underline.
func TableHeader(format string, args ...interface{}) {
	fmt.Printf("  "+format+"\n", args...)
	fmt.Printf("  %s\n", fmt.Sprintf(strings.Repeat("─", 100)))
}

// Column formats a column with fixed width.
func Column(val string, width int, color string) string {
	if len(val) > width {
		val = val[:width]
	}
	return fmt.Sprintf("%-*s", width, val)
}